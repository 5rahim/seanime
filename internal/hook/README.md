### `Request` events

- Route scoped
- A handler that does the native job is called last and can be interrupted if `e.next()` isn't called


### `Requested` events

- Example: `onAnimeEntryRequested`
- Called before creation of a struct
- Native job cannot be interrupted even if `e.next()` isn't called
- Followed by event containing the struct, e.g. `onAnimeEntry`
