import { buildSeaQuery } from "@/lib/server/queries/utils"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { ServerStatus } from "@/lib/server/types"

/**
 * Authentication
 * @param token
 */
export async function q_login(token: string) {
    return buildSeaQuery<ServerStatus, { token: string }>({
        endpoint: SeaEndpoints.LOGIN,
        method: "post",
        data: {
            token,
        },
    })
}