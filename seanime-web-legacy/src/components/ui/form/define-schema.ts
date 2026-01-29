import { z as zod, ZodType } from "zod"
import { schemaPresets } from "./schema-presets"

/* -------------------------------------------------------------------------------------------------
 * Helper type
 * -----------------------------------------------------------------------------------------------*/

export type InferType<S extends ZodType<any, any, any>> = zod.infer<S>

/* -------------------------------------------------------------------------------------------------
 * Helper functions
 * -----------------------------------------------------------------------------------------------*/

type DataSchemaCallback<S extends zod.ZodRawShape> = ({ z, presets }: {
    z: typeof zod,
    presets: typeof schemaPresets
}) => zod.ZodObject<S>

export const defineSchema = <S extends zod.ZodRawShape>(callback: DataSchemaCallback<S>): zod.ZodObject<S> => {
    return callback({ z: zod, presets: schemaPresets })
}
