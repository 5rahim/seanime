The database may store some structs defined outside as `bytes` inside `models`.
To avoid circular dependencies, we define methods that directly convert these `bytes` to the required struct using the database to store/retrieve them.
