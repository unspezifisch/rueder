# rueder

![rueder logo: a gopher with glasses reading a newspaper](docs/images/rueder3gopher_s.png)

## Dependencies

- Docker or podman. Everything is contained in docker images.
- docker-compose (1.27+ required)

## Installation

1. copy `rueder.example.env` to `rueder-prod.env` and
   1.1 change `LOGINSRV_JWT_SECRET` and `RUEDER_JWT` to something secure (eg. the output of `pwgen -s 128 1`)
   1.2 change `LOGINSRV_SIMPLE` to something else, however you want to authenticate your users
2. change `docker-compose.yaml` to match your setup's imgproxy URL
3. build docker production images: `docker-compose build`
4. start db only: `docker-compose up -d db`
5. create databases: `docker-compose -f docker-compose.initdb.yaml up`, then
   `Ctrl+C` after it's done
6. ready to run with: `docker-compose up -d`

## Development

While the `docker-compose.*` files in the main directory are for production use
the ones inside `backend/` and `frontend/` are intended for development purposes.
The Dockerfiles for production and development are named `Dockerfile` and
`Dockerfile.dev` respectively and are located next to each other in the `docker`
sub-directories inside the backend and frontend directories.

To start a local instance of `rueder3` with hot-reload on code changes, simply do:

```shell
# works in backend dir only:
make run

# OR

# works in backend and frontend dir:
docker-compose up --force-recreate --build
```

inside `backend/` directory in one terminal and inside `frontend/` directory in
another terminal. If you prefer to run the frontend without a docker container
you can also use `pnpm dev` instead.

**First run:** On the first run you need to initialize the database by running
`./utils/reset_db.sh` from the `backend/` directory. Restart all containers after this.

**Note:** Please use the provided Git hooks to make sure your changes pass
linting and testing.

### Services

The commands above spawn the following services, all listening on localhost:

- rueder web frontend that can be accessed at <http://127.0.0.1:3000> (with
  vite hot-reload)
- rueder backend processes (also with hot-reload): http feed api (:8080),
  authbackend (:8082, auth claims provider for loginsrv), events api (:8083),
  feedfinder api (:8081)
- additional required backend services: auth (:8082, loginsrv), postgres (:5432)
- additional utility services: imgproxy (:8086)

Open the web frontend and login with user `bob` and password `secret`.

You can also access the Swagger docs at <http://127.0.0.1:8080/swagger/index.html>
for the Feed API, and <http://127.0.0.1:8081/swagger/index.html> for the
Feedfinder.

### Git Hooks

#### Initial Setup for Development

Development Requirements:

- Python (3.7+ required) for pre-commit, pipx
- Docker for some pre-commit hooks
- Node/**p**npm for frontend checks (16+ required, 18+ recommended)
- Go (1.17+ required, 1.19 recommended) for backend checks (make sure `GOPATH` environment variable is set)
- golangci-lint for backend checks: <https://golangci-lint.run/usage/install/>

After installing these requirements and cloning the repository do these steps to
set up Git hooks:

```shell
# install pre-commit (mainly for backend stuff)
pipx install pre-commit
# note that something like ~/.local/bin needs to be in your PATH
pre-commit install

# install husky (mainly for frontend stuff)
cd frontend
pnpm install
# DO NOT run husky install
```

This installs [pre-commit](https://pre-commit.com/) which is triggered on Git commits.
The included config file `.pre-commit-config.yaml` then runs some backend checks
and finally runs [husky](https://typicode.github.io/husky) for frontend checks
using `frontend/.husky/`.

#### Running Manually

To just check if everything is in order:

```shell
pre-commit run --all-files
```

### Linting/Formatting

For backend use `gofmt` for formatting and the linters configured in `golangci-lint`.
Use `make lint` inside `./backend` directory to lint all files.

For frontend use (inside `./frontend` directory) `pnpm format` for formatting
and as linters use `pnpm lint` and `pnpm validate`.

### Testing

Look at the GitHub Workflows in `.github` if anything is unclear.

#### Backend

Use `make test` inside `./backend` to run tests using the included `_test.go` files.

When you change something which affects test execution in `APIPopRepository`
you need to start the test database service with `docker-compose up db` in `./backend`
and then run `make test_record` to update copyist's logfiles. The updated files
need to be committed, too.

#### Frontend

Use `pnpm test` to run tests with Jest using the included `.test.ts` files.

### Production Test

Before deploying to production you can test the production build of the
frontend with `pnpm preview` with the development version of the backend.

You can also test production backend and frontend with:

```shell
cp config/rueder.example.env config/rueder-test.env
# edit rueder-test.env to set LOGINSRV_JWT_SECRET and RUEDER_JWT values
docker network create http_default
docker-compose -f docker-compose.test.yaml up
```

The test-prod build will be accessible under <http://127.0.0.1:5000/rueder/>.
