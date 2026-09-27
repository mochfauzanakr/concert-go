package repository

type HelloRepository interface {
	PingDatabase() error
}
