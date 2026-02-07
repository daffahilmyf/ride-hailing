package outbound

type Tx interface {
	Commit() error
	Rollback() error
	RideRepo() RideRepo
	IdempotencyRepo() IdempotencyRepo
}

type TxManager interface {
	Begin() (Tx, error)
}
