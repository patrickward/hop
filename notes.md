# Notes

## Breaking changes in route branch

- Breaking: moved wrap middleware to `router` package
- Breaking: removed `wrap` package
- Breaking: removed route related methods on Server -- all route related methods are now on Router off of the app
- Breaking: the starting point for apps should be with `hop.New` instead of `hop.NewServer`
- Breaking: the `NewTemplateData` method moved from the server to the app
- Breaking: the `NewResponse` method moved from the server to the app
- New: added `hop.App` as a starting point with modules
- New: added `route` package for all routing and middleware
- New: added `hop.EventBus` for pub/sub internally
