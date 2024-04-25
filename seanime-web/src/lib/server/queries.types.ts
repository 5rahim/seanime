export type SeaErrorResponse = { error: string }
export type SeaDataResponse<T> = { data: T | undefined }
export type SeaResponse<T> = SeaDataResponse<T> | SeaErrorResponse
export type SeaWebsocketEvent<T> = { type: string, payload: T }
