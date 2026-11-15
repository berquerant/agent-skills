"""
Static AST analyzer for Python code snippets.

Reads Python source code from stdin, traverses the AST to extract cognitive
complexity metrics, line counts, and security hazard patterns (dynamic eval,
shell execution, and concurrent constructs), and outputs a JSON response to
stdout for ingestion by the Go snippet-analyzer feature.
"""

import ast
import json
import sys

CONTROL_FLOW_NODES = (
    ast.For,
    ast.AsyncFor,
    ast.While,
    ast.Try,
    ast.TryStar,
    ast.With,
    ast.AsyncWith,
    ast.Match,
)

ASYNC_MODULES = {"asyncio", "threading", "multiprocessing", "concurrent"}
SUBPROCESS_FUNCS = {"run", "call", "check_call", "check_output"}


class Analyzer:
    """AST visitor that computes complexity metrics and collects security hazards."""

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
        match node:
            case ast.If():
                self.walk_if(node, depth)
            case _ if isinstance(node, CONTROL_FLOW_NODES):
                self.walk_control_flow(node, depth + 1)
            case _:
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
        match node.orelse:
            case [ast.If() as elif_node]:
                self.walk(elif_node, depth)
            case orelse_stmts:
                for stmt in orelse_stmts:
                    self.walk(stmt, next_depth)

    def inspect_node(self, node):
        match node:
            case ast.FunctionDef() | ast.AsyncFunctionDef():
                self.record_function(node)
            case ast.If() | ast.IfExp() | ast.For() | ast.AsyncFor() | ast.While():
                self.record_loop_or_branch(node)
            case ast.Try() | ast.TryStar() | ast.Match() | ast.BoolOp() | ast.comprehension():
                self.record_complex_branch(node)
            case ast.AsyncWith() | ast.Await():
                self.metrics["background_count"] += 1
            case ast.Import() | ast.ImportFrom():
                self.inspect_import(node)
            case ast.Subscript(slice=ast.Constant(value=int() as idx)) if not isinstance(idx, bool):
                self.inspect_subscript(node, idx)
            case ast.ListComp() | ast.SetComp() | ast.DictComp() | ast.GeneratorExp():
                self.inspect_comprehension(node)
            case ast.Call():
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
        match node:
            case ast.Try(handlers=handlers) | ast.TryStar(handlers=handlers):
                self.metrics["branch_count"] += max(1, len(handlers))
            case ast.Match(cases=cases):
                self.metrics["branch_count"] += max(1, len(cases))
            case ast.BoolOp(values=values):
                self.metrics["branch_count"] += max(1, len(values) - 1)
            case ast.comprehension(ifs=ifs):
                self.metrics["branch_count"] += len(ifs)

    def inspect_import(self, node):
        match node:
            case ast.Import(names=aliases):
                for alias in aliases:
                    local = alias.asname or alias.name
                    self.aliases[local] = alias.name
                    if alias.name.split(".")[0] in ASYNC_MODULES:
                        self.metrics["background_count"] += 1
            case ast.ImportFrom(module=mod_name, names=aliases):
                mod = mod_name or ""
                if mod.split(".")[0] in ASYNC_MODULES:
                    self.metrics["background_count"] += 1
                for alias in aliases:
                    local = alias.asname or alias.name
                    self.aliases[local] = f"{mod}.{alias.name}" if mod else alias.name

    def inspect_subscript(self, node, idx):
        target = self.resolve_name(call_name(node.value))
        if target == "sys.argv" and idx > self.metrics["max_positional_arg"]:
            self.metrics["max_positional_arg"] = idx

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
        match name:
            case "eval" | "exec" | "compile" | "__import__" | "builtins.eval" | "builtins.exec":
                self.add_hazard("eval", line, f"dynamic code execution via {name}()", True)
            case "os.system" | "os.popen":
                self.metrics["subshell_count"] += 1
                self.add_hazard("pipe_to_shell", line, f"shell invocation via {name}()", True)
            case "importlib.import_module" | "runpy.run_path" | "runpy.run_module":
                self.add_hazard("source", line, f"dynamic module loading via {name}()", False)
            case "globals" | "locals" | "setattr":
                self.add_hazard("obscure_param_expansion", line, f"dynamic reflection via {name}()", False)
            case "signal.signal":
                self.metrics["trap_count"] += 1
            case _ if name.startswith("subprocess."):
                self.inspect_subprocess_call(node, name.split(".", 1)[1], line)
            case _ if name == "os.fork" or name.startswith("os.spawn"):
                self.metrics["background_count"] += 1

    def inspect_subprocess_call(self, node, attr, line):
        match attr:
            case "Popen":
                self.metrics["background_count"] += 1
            case _ if attr in SUBPROCESS_FUNCS:
                self.metrics["subshell_count"] += 1
        for kw in node.keywords:
            match kw:
                case ast.keyword(arg="shell", value=ast.Constant(value=True)):
                    self.add_hazard("pipe_to_shell", line, f"subprocess.{attr}(shell=True)", True)


def call_name(expr):
    match expr:
        case ast.Name(id=ident):
            return ident
        case ast.Attribute(value=val, attr=attr):
            prefix = call_name(val)
            return f"{prefix}.{attr}" if prefix else attr
        case _:
            return ""


def collect_docstring_lines(tree):
    doc_lines = set()
    for node in ast.walk(tree):
        match node:
            case (
                ast.Module(body=[ast.Expr(value=ast.Constant(value=str())) as first, *_])
                | ast.FunctionDef(body=[ast.Expr(value=ast.Constant(value=str())) as first, *_])
                | ast.AsyncFunctionDef(body=[ast.Expr(value=ast.Constant(value=str())) as first, *_])
                | ast.ClassDef(body=[ast.Expr(value=ast.Constant(value=str())) as first, *_])
            ):
                end_line = getattr(first, "end_lineno", first.lineno) or first.lineno
                doc_lines.update(range(first.lineno, end_line + 1))
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
