# Method Naming Conventions:

## For modifying single values:

- Use no prefix when setting a core property: `Layout()`, `Path()`, `Title()`
- Use `Status()` instead of `WithStatus()`

```go
// Before
resp.WithStatus(http.StatusOK).SetTitle("Page").Layout("base")

// After
resp.Status(http.StatusOK).Title("Page").Layout("base")
```

## For collections/maps:

- Use `Add` for appending to collections: `AddExtension()`, `AddHeader()`
- Use `Set` for replacing/overwriting: `SetData()`, `SetHeaders()`

```go
// Adding to collections
resp.AddExtension(NewHTMXExtension())
resp.AddHeader("Cache-Control", "no-cache")

// Setting/replacing
resp.SetData(newData)
resp.SetHeaders(headers)
```

## Status Method Conventions:

   For convenience methods, using the `Status` prefix makes the intent clearer:

```go
// Status convenience methods
func (t *TemplateResponse) StatusOK() *TemplateResponse {
    return t.Status(http.StatusOK)
}

func (t *TemplateResponse) StatusCreated() *TemplateResponse {
    return t.Status(http.StatusCreated)
}

// Usage becomes clear
resp.StatusOK().Path("users/list")
resp.StatusCreated().Path("users/show")
```
