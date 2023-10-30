# handlers

### ✅ Do

- Handle **database** logic.

### 🚫 Do not

- Do not define **helper functions** related to entities.

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

