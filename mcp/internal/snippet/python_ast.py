import ast
import json
import sys

TRY_NODES = (ast.Try, getattr(ast, "TryStar", ast.Try))
CONTROL_FLOW_NODES = (
    ast.For,
    ast.AsyncFor,
    ast.While,
    ast.With,
    ast.AsyncWith,
    ast.Match,
) + TRY_NODES

ASYNC_MODULES = {"asyncio", "threading", "multiprocessing", "concurrent"}
SUBPROCESS_FUNCS = {"run", "call", "check_call", "check_output"}


class Analyzer:
    def __init__(self):
        self.metrics = {
            "total_lines": 0,
            "sloc": 0,
            "comment_lines": 0,
            "function_count": 0,
            "max_function_lines": 0,
            "max_positional_arg": 0,
            "branch_count": 0,
            "max_nesting_depth": 0,
            "background_count": 0,
            "subshell_count": 0,
            "trap_count": 0,
            "max_pipeline_length": 0,
        }
        self.hazards = []
        self.aliases = {}

    def add_hazard(self, kind, line, detail, high_sev):
        self.hazards.append(
            {
                "kind": kind,
                "line": max(1, int(line or 1)),
                "detail": detail,
                "high_severity": bool(high_sev),
            }
        )

    def walk(self, node, depth=0):
        if depth > self.metrics["max_nesting_depth"]:
            self.metrics["max_nesting_depth"] = depth
        self.inspect_node(node)
        if isinstance(node, ast.If):
            self.walk_if(node, depth)
            return
        if isinstance(node, CONTROL_FLOW_NODES):
            self.walk_control_flow(node, depth + 1)
            return
        for child in ast.iter_child_nodes(node):
            self.walk(child, depth)

    def walk_control_flow(self, node, next_depth):
        if next_depth > self.metrics["max_nesting_depth"]:
            self.metrics["max_nesting_depth"] = next_depth
        for child in ast.iter_child_nodes(node):
            self.walk(child, next_depth)

    def walk_if(self, node, depth):
        next_depth = depth + 1
        if next_depth > self.metrics["max_nesting_depth"]:
            self.metrics["max_nesting_depth"] = next_depth
        self.walk(node.test, depth)
        for stmt in node.body:
            self.walk(stmt, next_depth)
        if len(node.orelse) == 1 and isinstance(node.orelse[0], ast.If):
            self.walk(node.orelse[0], depth)
            return
        for stmt in node.orelse:
            self.walk(stmt, next_depth)

    def inspect_node(self, node):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            self.record_function(node)
        elif isinstance(node, (ast.If, ast.IfExp, ast.For, ast.AsyncFor, ast.While)):
            self.record_loop_or_branch(node)
        elif isinstance(node, TRY_NODES + (ast.Match, ast.BoolOp, ast.comprehension)):
            self.record_complex_branch(node)
        elif isinstance(node, (ast.AsyncWith, ast.Await)):
            self.metrics["background_count"] += 1
        elif isinstance(node, (ast.Import, ast.ImportFrom)):
            self.inspect_import(node)
        elif isinstance(node, ast.Subscript):
            self.inspect_subscript(node)
        elif isinstance(node, (ast.ListComp, ast.SetComp, ast.DictComp, ast.GeneratorExp)):
            self.inspect_comprehension(node)
        elif isinstance(node, ast.Call):
            self.inspect_call(node)

    def record_function(self, node):
        self.metrics["function_count"] += 1
        end_line = getattr(node, "end_lineno", node.lineno) or node.lineno
        span = end_line - node.lineno + 1
        if span > self.metrics["max_function_lines"]:
            self.metrics["max_function_lines"] = span
        pos_args = len(node.args.posonlyargs) + len(node.args.args)
        if pos_args > self.metrics["max_positional_arg"]:
            self.metrics["max_positional_arg"] = pos_args
        if isinstance(node, ast.AsyncFunctionDef):
            self.metrics["background_count"] += 1

    def record_loop_or_branch(self, node):
        self.metrics["branch_count"] += 1
        if isinstance(node, ast.AsyncFor):
            self.metrics["background_count"] += 1

    def record_complex_branch(self, node):
        if isinstance(node, TRY_NODES):
            self.metrics["branch_count"] += max(1, len(node.handlers))
        elif isinstance(node, ast.Match):
            self.metrics["branch_count"] += max(1, len(node.cases))
        elif isinstance(node, ast.BoolOp):
            self.metrics["branch_count"] += max(1, len(node.values) - 1)
        elif isinstance(node, ast.comprehension):
            self.metrics["branch_count"] += len(node.ifs)

    def inspect_import(self, node):
        if isinstance(node, ast.Import):
            for alias in node.names:
                local = alias.asname or alias.name
                self.aliases[local] = alias.name
                if alias.name.split(".")[0] in ASYNC_MODULES:
                    self.metrics["background_count"] += 1
            return
        mod = node.module or ""
        if mod.split(".")[0] in ASYNC_MODULES:
            self.metrics["background_count"] += 1
        for alias in node.names:
            local = alias.asname or alias.name
            self.aliases[local] = f"{mod}.{alias.name}" if mod else alias.name

    def inspect_subscript(self, node):
        target = self.resolve_name(call_name(node.value))
        if target != "sys.argv":
            return
        idx = node.slice
        if isinstance(idx, ast.Constant) and isinstance(idx.value, int) and not isinstance(idx.value, bool):
            if idx.value > self.metrics["max_positional_arg"]:
                self.metrics["max_positional_arg"] = idx.value

    def inspect_comprehension(self, node):
        if len(node.generators) >= 2:
            self.add_hazard(
                "obscure_param_expansion",
                node.lineno,
                f"multi-generator comprehension ({len(node.generators)} clauses)",
                False,
            )

    def resolve_name(self, raw):
        if not raw:
            return ""
        parts = raw.split(".")
        if parts[0] in self.aliases:
            parts[0] = self.aliases[parts[0]]
        return ".".join(parts)

    def inspect_call(self, node):
        name = self.resolve_name(call_name(node.func))
        line = getattr(node, "lineno", 1)
        if name in {"eval", "exec", "compile", "__import__", "builtins.eval", "builtins.exec"}:
            self.add_hazard("eval", line, f"dynamic code execution via {name}()", True)
        elif name in {"os.system", "os.popen"}:
            self.metrics["subshell_count"] += 1
            self.add_hazard("pipe_to_shell", line, f"shell invocation via {name}()", True)
        elif name in {"importlib.import_module", "runpy.run_path", "runpy.run_module"}:
            self.add_hazard("source", line, f"dynamic module loading via {name}()", False)
        elif name in {"globals", "locals", "setattr"}:
            self.add_hazard("obscure_param_expansion", line, f"dynamic reflection via {name}()", False)
        elif name == "signal.signal":
            self.metrics["trap_count"] += 1
        elif name.startswith("subprocess."):
            self.inspect_subprocess_call(node, name.split(".", 1)[1], line)
        elif name == "os.fork" or name.startswith("os.spawn"):
            self.metrics["background_count"] += 1

    def inspect_subprocess_call(self, node, attr, line):
        if attr == "Popen":
            self.metrics["background_count"] += 1
        elif attr in SUBPROCESS_FUNCS:
            self.metrics["subshell_count"] += 1
        for kw in node.keywords:
            if kw.arg == "shell" and isinstance(kw.value, ast.Constant) and kw.value.value is True:
                self.add_hazard("pipe_to_shell", line, f"subprocess.{attr}(shell=True)", True)


def call_name(expr):
    if isinstance(expr, ast.Name):
        return expr.id
    if isinstance(expr, ast.Attribute):
        prefix = call_name(expr.value)
        return f"{prefix}.{expr.attr}" if prefix else expr.attr
    return ""


def collect_docstring_lines(tree):
    doc_lines = set()
    for node in ast.walk(tree):
        if not isinstance(node, (ast.Module, ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
            continue
        if not getattr(node, "body", None):
            continue
        first = node.body[0]
        if isinstance(first, ast.Expr) and isinstance(first.value, ast.Constant) and isinstance(first.value.value, str):
            end_line = getattr(first, "end_lineno", first.lineno) or first.lineno
            for ln in range(first.lineno, end_line + 1):
                doc_lines.add(ln)
    return doc_lines


def compute_line_metrics(source, tree):
    normalized = source.replace("\r\n", "\n").rstrip("\n")
    if not normalized.strip():
        return 0, 0, 0
    lines = normalized.split("\n")
    doc_lines = collect_docstring_lines(tree)
    sloc = 0
    comments = 0
    for idx, raw in enumerate(lines, start=1):
        stripped = raw.strip()
        if not stripped:
            continue
        if stripped.startswith("#") or idx in doc_lines:
            comments += 1
        else:
            sloc += 1
    return len(lines), sloc, comments


def main():
    source = sys.stdin.read()
    try:
        tree = ast.parse(source, filename="snippet.py")
    except SyntaxError as exc:
        msg = f"SyntaxError: {exc.msg} (line {exc.lineno})"
        print(json.dumps({"ok": False, "error": msg}))
        return
    except Exception as exc:
        print(json.dumps({"ok": False, "error": f"{type(exc).__name__}: {exc}"}))
        return

    analyzer = Analyzer()
    analyzer.walk(tree, 0)
    total_lines, sloc, comments = compute_line_metrics(source, tree)
    analyzer.metrics["total_lines"] = total_lines
    analyzer.metrics["sloc"] = sloc
    analyzer.metrics["comment_lines"] = comments
    print(json.dumps({"ok": True, "metrics": analyzer.metrics, "hazards": analyzer.hazards}))


if __name__ == "__main__":
    main()
