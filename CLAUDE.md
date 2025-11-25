## Conventions

We're building a 12 Factor app.

Logging is important. Especially in this stage as we're going to need to debug things. Take advantage of charmbracelet/log's methods.

## Dependencies

This project uses github.com/charmbracelet/log instead of go std log.
Logging values need to be prefixed with a string name e.g. `log.Info("My log", "value", value)`.
