import { z } from "zod"

export const schemaPresets = {
    name: z.string().min(2).trim(),
    select: z.string().min(1),
    checkboxGroup: z.array(z.string()),
    multiSelect: z.array(z.string()),
    autocomplete: z.object({ label: z.string(), value: z.string().nullable() }),
    validAddress: z.object({
        label: z.string(), value: z.string({
            required_error: "Invalid address",
            invalid_type_error: "Invalid address",
        }),
    }),
    time: z.object({ hour: z.number().min(0).max(23), minute: z.number().min(0).max(59) }),
    phone: z.string().min(10, "Invalid phone number"),
    files: z
        .array(z.custom<File>())
        .refine((files) => files.every((file) => file instanceof File), { message: "Expected a file" }),
    filesOrEmpty: z
        .array(z.custom<File>()).min(0)
        .refine((files) => files.every((file) => file instanceof File), { message: "Expected a file" }),
    dateRangePicker: z.object({ from: z.date(), to: z.date() }),
    datePicker: z.date(),
}
