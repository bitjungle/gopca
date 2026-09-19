# GoPCA model schemas

JSON Schema definitions for the model files GoPCA writes and reads.

These are published as a **format specification**. Every model file GoPCA
exports carries a `$schema` pointer into this directory, so validators and
editors can resolve it:

```json
"$schema": "https://raw.githubusercontent.com/bitjungle/gopca/main/schemas/v2/pca-output.schema.json"
```

`v2` is current. `v1` is retained so that older model files remain readable.

The schemas describe the file format only. They are published so that GoPCA
model files can be read, validated, and written by other tools without needing
access to the GoPCA source code.
