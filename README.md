# bank - architectural experiment project

Note: Developments are ongoing. This is not the final version.

This is a simple bank application that allows you to create accounts, deposit and withdraw money, and transfer money between accounts.

But the main goal of this project is to experiment with different architectural patterns and technologies. The project is uses clean architecture with golang standarts. Also this project uses some technologies like opentelemetry, grpc, prometheus, jaeger, kubernetes, etc.

When we get exactly the structure we want, we will separate its bones and move it to [gopre's repository](https://github.com/9ssi7/gopre).

## Useful Links

- [`golang code review comments`](https://go.dev/wiki/CodeReviewComments)
- [`effective go docs`](https://go.dev/doc/effective_go)
- [`golang styleguide by Google`](https://google.github.io/styleguide/go/decisions)

## Run App

- `make once` - Run the app once for jwt secret key generation and docker network creation.
- `make compose` - Run the app with docker-compose for dependencies.
- `make build-srv && make start-srv` - Build and Run the app.