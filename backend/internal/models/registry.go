package models

func ModelRegistry() []interface{} {
	return []interface{}{
		&User{},
		&SystemLog{},
		&Listmak{},
		&Order{},
		&ShareLink{},
		&ViewShare{},
		&AILog{},
		&PriceCatalog{},
		&ListmakSummary{},
		&Payment{},
		&PaymentLog{},
		&AppSetting{},
	}
}
