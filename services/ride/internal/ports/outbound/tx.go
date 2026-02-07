package outbound

type Tx interface {
	Commit() error
	Rollback() error
	RideRepo() RideRepo
	IdempotencyRepo() IdempotencyRepo
	OutboxRepo() OutboxRepo
	RideOfferRepo() RideOfferRepo
}

type TxManager interface {
	Begin() (Tx, error)
}
