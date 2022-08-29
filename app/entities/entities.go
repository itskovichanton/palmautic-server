package entities

type ID int64

type IBaseEntity interface {
	GetId() ID
	GetAccountId() ID
}

type BaseEntity struct {
	IBaseEntity
	Id, AccountId ID
}

func (c BaseEntity) GetId() ID {
	return c.Id
}

func (c BaseEntity) GetAccountId() ID {
	return c.AccountId
}

type Contact struct {
	BaseEntity
	Phone, Name, Email string
}
