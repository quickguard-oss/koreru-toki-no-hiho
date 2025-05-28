/*
Package cmd provides the command-line interface functionality for the application.

It implements various commands, flags, and their handlers using the cobra library.
This package serves as the entry point for all CLI operations and coordinates the
execution flow between different components of the application.

Each command is organized into its own file and registered with the root command.
Command handlers parse input flags, validate user input, and delegate to the
appropriate business logic in other packages.
*/
package cmd
