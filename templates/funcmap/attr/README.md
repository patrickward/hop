### `attr.FuncMap` Template Functions

The `attr.FuncMap` provides several utility functions for working with HTML attributes in Go templates. Below is a description of each function and examples of how they can be used in templates.

#### `attr_class`

Combines multiple class names into a single string, filtering out any empty strings.

**Usage:**

```gotemplate
<div class="{{ attr_class "card" (when .IsLarge "card-lg") (unless .IsVisible "hidden") }}">Content</div>
```

#### `attr_safe`

Marks a string as safe for HTML attribute output.

**Usage:**

```gotemplate
<input type="text" value="{{ attr_safe .Value }}">
```

#### `attr_selected`

Returns "selected" if the `current` value matches the `value`, otherwise returns an empty string.

**Usage:**

```gotemplate
<option value="option1" {{ attr_selected .CurrentValue "option1" }}>Option 1</option>
```

#### `attr_checked`

Returns "checked" if the `current` value matches the `value`, otherwise returns an empty string.

**Usage:**

```gotemplate
<input type="checkbox" {{ attr_checked .IsChecked "true" }}>
```

#### `attr_disabled`

Returns "disabled" if the `current` value matches the `value`, otherwise returns an empty string.

**Usage:**

```gotemplate
<button {{ attr_disabled .IsDisabled "true" }}>Submit</button>
```

#### `attr_readonly`

Returns "readonly" if the `current` value matches the `value`, otherwise returns an empty string.

**Usage:**

```gotemplate
<input type="text" {{ attr_readonly .IsReadonly "true" }}>
```

These functions help ensure that HTML attributes are correctly and safely rendered in templates, improving both security and readability.
