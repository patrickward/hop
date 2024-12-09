# Mail Package

The mail package provides a simple yet flexible email sending solution for Go applications using SMTP. It supports HTML and plain text emails, attachments, templates, and retries.

## Features

- HTML and plain text email support
- Go template-based email composition
- File attachments with content type detection
- Configurable retry mechanism
- HTML processing (optional)
- CC, BCC, and Reply-To support
- Multiple recipients
- SMTP authentication

## Installation

```bash
go get github.com/patrickward/hop/mail
```

## Quick Start

```go
// Create configuration
config := &mail.Config{
    Host:     "smtp.example.com",
    Port:     587,
    Username: "user@example.com",
    Password: "password",
    From:     "sender@example.com",
}

// Create mailer
mailer, err := mail.NewMailer(config)
if err != nil {
    log.Fatal(err)
}

// Send an email
msg, err := mail.NewMessage().
    To("recipient@example.com").
    Template("emails/welcome.tmpl").
    WithData(map[string]any{
        "name": "John",
    }).
    Build()

if err != nil {
    log.Fatal(err)
}

err = mailer.Send(msg)
```

## Configuration

```go
type Config struct {
    // SMTP Configuration
    Host      string        // SMTP server host
    Port      int          // SMTP server port
    Username  string       // SMTP authentication username
    Password  string       // SMTP authentication password
    From      string       // Default sender address
    AuthType  string       // Auth type (e.g., "LOGIN", "PLAIN", "NOAUTH")
    
    // Template Configuration
    TemplateFS    fs.FS    // Filesystem for templates
    TemplatePath  string   // Optional base path for templates
    
    // Retry Configuration
    RetryCount    int          // Number of retry attempts
    RetryDelay    time.Duration // Delay between retries
    
    // Optional HTML Processing
    HTMLProcessor HTMLProcessor // Optional HTML processor
}
```

## Email Templates

Templates must define three sections:
- `subject`: The email subject
- `text/plain`: Plain text version of the email
- `text/html`: HTML version of the email (optional)

Example template:

```go
{{define "subject"}}Welcome to {{.Company}}{{end}}

{{define "text/plain"}}
Hello {{.Name}},

Welcome to {{.Company}}! We're glad to have you on board.

Best regards,
The Team
{{end}}

{{define "text/html"}}
<html>
<body>
    <h1>Welcome to {{.Company}}</h1>
    <p>Hello {{.Name}},</p>
    <p>We're glad to have you on board.</p>
    <p>Best regards,<br>The Team</p>
</body>
</html>
{{end}}
```

## Working with Attachments

### Basic Attachment
```go
content := strings.NewReader("Hello World")
msg, err := mail.NewMessage().
    To("recipient@example.com").
    Template("template.tmpl").
    Attach("hello.txt", content).
    Build()
```

### File Attachment
```go
filename, reader, cleanup, err := mail.OpenFileAttachment("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer cleanup()

msg, err := mail.NewMessage().
    To("recipient@example.com").
    Template("template.tmpl").
    Attach(filename, reader).
    Build()
```

### Custom Content Type
```go
msg, err := mail.NewMessage().
    To("recipient@example.com").
    Template("template.tmpl").
    AttachWithContentType("page.html", reader, gomail.TypeTextHTML).
    Build()
```

## Advanced Usage

### Multiple Recipients
```go
msg, err := mail.NewMessage().
    To("recipient1@example.com", "recipient2@example.com").
    Cc("cc@example.com").
    Bcc("bcc@example.com").
    ReplyTo("reply@example.com").
    Template("template.tmpl").
    Build()
```

### Multiple Templates
```go
msg, err := mail.NewMessage().
    To("recipient@example.com").
    Template(
        "emails/header.tmpl",
        "emails/content.tmpl",
    ).
    Build()
```

## HTML Processing

The package supports custom HTML processing through the HTMLProcessor interface:

```go
type HTMLProcessor interface {
    Process(html string) (string, error)
}
```

This can be used for tasks like CSS inlining or HTML modification before sending.

## Known Limitations

1. Template Requirements
    - Templates must provide at least `subject` and `text/plain` sections
    - HTML body (`text/html`) is optional but recommended

2. SMTP Support
    - Limited to SMTP protocol
    - No direct support for API-based email services

3. Attachments
    - Files must be readable at send time
    - No automatic MIME type detection
    - File handles must be managed properly

4. HTML Processing
    - Basic HTML support only
    - No built-in CSS inlining
    - Custom processing requires implementing HTMLProcessor interface

5. Templates
    - All templates must be available in the provided TemplateFS
    - No dynamic template loading
    - No support for external template sources

## Best Practices

1. Template Management
    - Keep templates organized in a dedicated directory
    - Use consistent naming conventions
    - Include both HTML and plain text versions

2. Error Handling
    - Always check Build() errors before sending
    - Implement proper retry handling for temporary failures
    - Log send failures appropriately

3. Resource Management
    - Always use cleanup functions when working with file attachments
    - Close file handles after sending
    - Monitor SMTP connection status

4. Security
    - Use TLS when possible
    - Protect credentials
    - Validate email addresses
    - Sanitize template data

## Dependencies

- [github.com/wneessen/go-mail](https://github.com/wneessen/go-mail) - SMTP client
- Standard library packages only for core functionality

## Testing

The package includes both unit tests and integration tests. Integration tests require a running SMTP server (like MailPit) and can be enabled by setting the `TEST_MAILPIT=1` environment variable.

