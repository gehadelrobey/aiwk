## Example files (for testing)

This repo includes a few tiny inputs under `examples/` so you can quickly try commands:

```bash
cat examples/columns.txt | aiwk "print the first and second columns"
cat examples/passwd.txt | aiwk -F: "print the username and shell"
cat examples/sales.csv | aiwk --csv -F, "sum amount_usd by customer"
cat examples/access.log | aiwk "sum bytes (field 10) grouped by IP (field 1)"
```