# Examples

Small, focused examples — each in its own subdirectory with a single
`main.go`. Run any of them with `go run ./examples/<name>`. Press
`Ctrl+Q` to quit.

| Example | Demonstrates |
|---------|--------------|
| [form](./form) | TextInput + Button + Field, `OnSubmit` / `OnPress`, global key bindings |
| [markdown](./markdown) | `MarkdownTextArea` rendering headings, bold, code, links (OSC 8) |
| [popup](./popup) | `PopupWindow` with `Close(result)` and `OnResult` callback |
| [table](./table) | `Table` with `SliceDataSource`, flex column, per-column `CellRenderer` |

For a richer playground covering more components in one app, see
[`uidemo/`](../uidemo) and [`tabdemo/`](../tabdemo).
