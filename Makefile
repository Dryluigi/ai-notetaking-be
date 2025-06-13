create-migration:
	migrate create -ext sql -dir migrations $(name)

migrate-db:
	migrate -path migrations -database "postgres://user:password@localhost:5435/ai-notetaking?sslmode=disable" up