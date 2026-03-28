[![Go Report Card](https://goreportcard.com/badge/github.com/rwrrioe/sso)](https://goreportcard.com/report/github.com/rwrrioe/sso)&nbsp;[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub tag](https://img.shields.io/github/v/tag/rwrrioe/sso?label=version)](https://github.com/rwrrioe/sso/releases)

## About

SSO service is a lightweight gRPC SSO authentication / authorization provider. Fully written in Golang. 

- Login / Register provider
- IsAdmin, other roles provider (authz)
- Password reset, email verification
- Cookies, session, logout (to be added)
  


## Quick start

1. Do git clone
```
git clone github.com/rwrrioe/sso
```

2. Install [Installation | Task](https://taskfile.dev/docs/installation) 

3. Build container
```
docker build
```

---

## Architecture

## Authentication pipeline

![Architecture](docs/assets/ssoArchitecture.svg)


Our gRPC server calls Auth usecase. If it's register request, Auth calls repository and inserts credentials. We should add salt to our passHash before insertion. If it's login request, Auth compares hashes from user credentials and User DB. 


