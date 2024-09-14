The database may store some structs defined outside as `[]byte` inside `models`.
To avoid circular dependencies, we define methods that directly convert `[]byte` to the corresponding struct using the database to store/retrieve them.
