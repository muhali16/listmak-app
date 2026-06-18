package models

func ModelRegistry() []interface{} {
	return []interface{}{
		&User{},
		&SystemLog{},
		&Listmak{},
		&Order{},
		&ShareLink{},
		&ViewShare{},
	}
}
