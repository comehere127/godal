# godal

[![Go Report Card](https://goreportcard.com/badge/github.com/btnguyen2k/godal)](https://goreportcard.com/report/github.com/btnguyen2k/godal)
[![GoDoc](https://godoc.org/github.com/btnguyen2k/godal?status.svg)](https://godoc.org/github.com/btnguyen2k/godal)

Generic Database Access Layer implementation in Golang.

## Documentation

- [GoDoc](https://godoc.org/github.com/btnguyen2k/godal)
- [Examples](examples/)
- [Generic MongoDB DAO](mongo/MONGO.md)
- [Generic database/sql DAO](sql/SQL.md)


## History

### 2019-04-09 - v0.0.2

- Migrate to Go modular design.
- Generic [database/sql](https://golang.org/pkg/database/sql/) DAO implementation: `GenericDaoSql`


### 2019-03-12 - v0.0.1

- Generic business object:
  - Interface `IGenericBo` & implementation `GenericBo`
- Generic data access object:
  - Interface `IGenericDao` & abstract implementation `AbstractGenericDao`
- Generic [MongoDB](https://www.mongodb.com) DAO implementation: `GenericDaoMongo`


## License

MIT - see [LICENSE.md](LICENSE.md).