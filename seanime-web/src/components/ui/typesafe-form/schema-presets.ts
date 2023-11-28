import _isDate from "lodash/isDate"
import { z } from "zod"

/**
 * @internal
 * createTypesafeFormSchema presets
 */
export const schemaPresets = {
    name: z.string().min(2).trim(),
    select: z.string().nonempty(),
    checkboxGroup: z.array(z.string()),
    multiSelect: z.array(z.string()),
    radioGroup: z.string().nonempty(),
    dropzone: z.array(z.custom<File>()).refine(
        // Check if all items in the array are instances of the File object
        (files) => files.every((file) => file instanceof File), { message: "Expected a file" },
    ),
    time: z.object({ hour: z.number().min(0).max(23), minute: z.number().min(0).max(59) }),
    phone: z.string().min(10, "Invalid phone number"),
    price: z.number().min(0),
    switch: z.boolean(),
    checkbox: z.boolean(),
    files: z.array(z.custom<File>()).refine(
        // Check if all items in the array are instances of the File object
        (files) => files.every((file) => file instanceof File), { message: "Expected a file" },
    ),
    dateRangePicker: z.object({ start: z.date(), end: z.date() })
        .refine(data => _isDate(data.start) && _isDate(data.end), { message: "Incorrect dates" }),
    datePicker: z.date(),
}
