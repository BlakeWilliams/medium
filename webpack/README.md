# Webpack Middleware

This package implements a basic webpack dev server middleware, useful for
serving assets in the development environment.

## Usage

TODO: In the meantime, see webpack_test.go

## Development

This package uses a real instance of the webpack dev server for a complete
integration test. To ensure tests will pass, run `npm install` and validate
that `npx webpack serve` works within the `test_env` directory.

## Contributing

Bug fixes and minor improvements are welcome in the form of Pull Requests. For
larger changes please open issue to discuss the proposed change first.
