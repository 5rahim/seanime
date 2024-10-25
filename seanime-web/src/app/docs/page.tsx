"use client"


import { DatePicker } from "@/components/ui/date-picker"
import { normalizeDate } from "@/lib/helpers/date"
import React from "react"

export default function Page() {

    const [value, setValue] = React.useState<Date | undefined>(undefined)

    React.useEffect(() => {

        setValue(normalizeDate("2024-04-15T00:00:00Z"))
        console.log(normalizeDate("2024-04-15T00:00:00Z"))
        console.log(new Date("2024-04-15T00:00:00Z"))

    }, [])

    React.useEffect(() => {
        console.log(value)
    }, [value])


    return (
        <div className="space-y-4 container py-10">
            <DatePicker
                value={value}
                onValueChange={setValue}
            />
            {/*{data?.routeGroups?.map((group, i) => (*/}
            {/*    <div key={group.filename + i} className="space-y-4">*/}
            {/*        <h4 className="">{group.filename}</h4>*/}

            {/*        <Accordion type="multiple" defaultValue={[]}>*/}
            {/*            {group.routes.toSorted((a, b) => a.endpoint.length - b.endpoint.length).map((route, i) => (*/}
            {/*                <AccordionItem value={route.name} key={route.name + i} className="space-y-2">*/}
            {/*                    <AccordionTrigger className="rounded flex-none w-full">*/}
            {/*                        <p className="flex gap-2 items-center">*/}
            {/*                            <Badge*/}
            {/*                                className="w-24 py-4"*/}
            {/*                                intent={(route.methods.includes("GET") && route.methods.length === 1) ? "success"*/}
            {/*                                    : route.methods.includes("GET") ? "warning"*/}
            {/*                                        : route.methods.includes("DELETE") ? "alert"*/}
            {/*                                            : route.methods.includes("PATCH") ? "warning" : "primary"}*/}
            {/*                            >*/}
            {/*                                {route.methods.join(", ")}*/}
            {/*                            </Badge>*/}
            {/*                            <span className="font-semibold flex-none whitespace-nowrap">{route.endpoint}</span>*/}
            {/*                            /!*<span className="font-medium text-[--muted] text-sm truncate flex-shrink">({route.name.replace("Handle", "")})</span>*!/*/}
            {/*                            <span className="text-[--muted] text-[.97rem] whitespace-nowrap truncate text-ellipsis"> - {route.summary}</span>*/}
            {/*                        </p>*/}
            {/*                    </AccordionTrigger>*/}

            {/*                    <AccordionContent className="space-y-4 bg-gray-50 border rounded mb-4">*/}
            {/*                        <p className="font-bold">*/}
            {/*                            {route.name.replace("Handle", "")}*/}
            {/*                        </p>*/}
            {/*                        {!!route.description && <p className="">{route.description}</p>}*/}

            {/*                        <div className="space-y-2">*/}
            {/*                            <h5>Params</h5>*/}
            {/*                            <ul className="list-disc pl-4">*/}
            {/*                                {route.params.map((param, i) => (*/}
            {/*                                    <li key={param.name + i} className="flex gap-2 items-center">*/}
            {/*                                        <p className="font-medium">*/}
            {/*                                            {param.name}*/}
            {/*                                            {param.required && <span className="text-red-500">*</span>}*/}
            {/*                                        </p>*/}
            {/*                                        <p className="text-[--muted]">{param.type}</p>*/}
            {/*                                        <p>{param.description}</p>*/}
            {/*                                    </li>*/}
            {/*                                ))}*/}
            {/*                            </ul>*/}
            {/*                        </div>*/}

            {/*                        <div className="space-y-2">*/}
            {/*                            <h5>Request Body Fields</h5>*/}
            {/*                            <ul className="list-disc pl-4">*/}
            {/*                                {route.requestBodyFields.map((field, i) => (*/}
            {/*                                    <li key={field.name + i} className="flex gap-2 items-center">*/}
            {/*                                        <p className="font-medium">{field.name}</p>*/}
            {/*                                        <p className="text-[--muted]">{field.type}</p>*/}
            {/*                                        <p>{field.description}</p>*/}
            {/*                                    </li>*/}
            {/*                                ))}*/}
            {/*                            </ul>*/}
            {/*                        </div>*/}

            {/*                        <div className="flex gap-2 items-center">*/}
            {/*                            <p className="font-medium text-[--muted]">Returns</p>*/}
            {/*                            <p className="font-bold text-brand-900">{route.returns}</p>*/}
            {/*                        </div>*/}
            {/*                    </AccordionContent>*/}
            {/*                </AccordionItem>*/}
            {/*            ))}*/}
            {/*        </Accordion>*/}

            {/*        <Separator />*/}
            {/*    </div>*/}
            {/*))}*/}
        </div>
    )
}
