package dbulker

type Database interface {
	Connect() error
	FindAll(string, interface{}) ([]interface{}, error)
	FillAutoPlainMultiple(string, []interface{}) error
	FillAutoPlainSingle(string, interface{}) (int64, error)
}
