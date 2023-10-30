## Routes

## /api/v1

### `POST` /api/v1/auth

- Saves the username and token in the database (account)

#### Request

```json
{
  "token": "string"
}
```

#### Response

`TODO`

### `POST` /api/v1/scan

- Requires authentication

#### Request

```json
{
  "enhanced": "boolean"
}
```

