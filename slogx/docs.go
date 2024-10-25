// Package slogx provides some utility methods based around the standard library's slog.Logger package.
// The goal is to make it easier to create a new slog.Logger given a set of options. More specfically,
// within the context of the project, we want to be able to create a new slogx based on the config
// options provided by the user via environment variables or the CLI.
package slogx
