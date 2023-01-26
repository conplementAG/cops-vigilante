package database

type InMemoryDatabase struct {
	store map[string]interface{}
}

func NewInMemoryDatabase() Database {
	return &InMemoryDatabase{store: map[string]interface{}{}}
}

func (db *InMemoryDatabase) Get(key string) interface{} {
	return db.store[key]
}

func (db *InMemoryDatabase) GetAll() map[string]interface{} {
	return db.store
}

func (db *InMemoryDatabase) Set(key string, value interface{}) {
	db.store[key] = value
}

func (db *InMemoryDatabase) Delete(key string) {
	delete(db.store, key)
}
