# Diff Continuation Plan Template

Use this template in Step 4 of `diff-continue` to present the inferred transformation rule, already modified files, remaining target locations, and planned verification before applying edits.

```markdown
### Diff Continuation Plan

#### Inferred Transformation Rule
<Describe the exact rule, e.g. "Add ctx context.Context as first parameter to all exported handler functions and propagate to caller.">

#### Already Changed Files
- `pkg/api/user.go`
- `pkg/api/auth.go`

#### Remaining Target Files
- `pkg/api/order.go` (methods: `CreateOrder`, `GetOrder`)
- `pkg/api/payment.go` (methods: `ProcessPayment`)

#### Ambiguous / Excluded Scope (if any)
- `pkg/api/*_test.go`: <Confirm whether unit test signatures should also be updated in this pass>

#### Planned Verification
- Run tests: `<test command>`
- Inspect full diff: `git diff HEAD`

May I proceed with applying this pattern to the remaining targets?
```
