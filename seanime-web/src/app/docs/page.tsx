"use client"

import { useGetDocs } from "@/api/hooks/docs.hooks"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import React from "react"

export default function Page() {

    const { data, isLoading } = useGetDocs()

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4 container py-10">
            {data?.toSorted?.((a, b) => a.filename?.localeCompare(b.filename))?.map((group, i) => (
                <div key={group.filename + i} className="space-y-4">
                    <h4 className=""><span>{group.filename}</span> <span className="text-gray-300">/</span>
                        <span className="text-[--muted]"> {group.filename.replace(".go", "")}.hooks.ts</span></h4>
                    <Accordion type="multiple" defaultValue={[]}>
                        {group.handlers?.toSorted((a, b) => a.filename?.localeCompare(b.filename)).map((route, i) => (
                            <AccordionItem value={route.name} key={route.name + i} className="space-y-2">
                                <AccordionTrigger className="rounded flex-none w-full">
                                    <p className="flex gap-2 items-center">
                                        <Badge
                                            className="w-24 py-4"
                                            intent={(route.api!.methods?.includes("GET") && route.api!.methods?.length === 1) ? "success"
                                                : route.api!.methods?.includes("GET") ? "warning"
                                                    : route.api!.methods?.includes("DELETE") ? "alert"
                                                        : route.api!.methods?.includes("PATCH") ? "warning" : "primary"}
                                        >
                                            {route.api!.methods?.join(", ")}
                                        </Badge>
                                        <span className="font-semibold flex-none whitespace-nowrap">{route.api!.endpoint}</span>
                                        <span className="font-normal text-sm text-[--muted] flex-none whitespace-nowrap">{route.name}</span>
                                        {/*<span className="font-medium text-[--muted] text-sm truncate flex-shrink">({route.name.replace("Handle", "")})</span>*/}
                                        <span className="text-[--muted] text-[.97rem] whitespace-nowrap truncate text-ellipsis"> - {route.api!.summary}</span>
                                    </p>
                                </AccordionTrigger>

                                <AccordionContent className="space-y-4 border rounded mb-4">
                                    {/*<p className="font-bold">*/}
                                    {/*    {route.name}*/}
                                    {/*</p>*/}
                                    {/*<p className="">*/}
                                    {/*    Used in: <span className="font-bold">{route.filename.replace(".go", "")}.hooks.ts</span>*/}
                                    {/*</p>*/}
                                    {!!route.api!.descriptions?.length && <div>
                                        {route.api!.descriptions?.map((desc, i) => (
                                            <p key={desc + i}>{desc}</p>
                                        ))}
                                    </div>}

                                    {!!route.api!.params?.length && <div className="space-y-2">
                                        <h5>URL Params</h5>
                                        <ul className="list-disc pl-4">
                                            {route.api!.params?.map((param, i) => (
                                                <li key={param.name + i} className="flex gap-2 items-center">
                                                    <p className="font-medium">
                                                        {param.name}
                                                        {param.required && <span className="text-red-500">*</span>}
                                                    </p>
                                                    <p className="text-[--muted]">{param.typescriptType}</p>
                                                    {param.descriptions?.map((desc, i) => (
                                                        <p key={desc + i}>{desc}</p>
                                                    ))}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>}

                                    {!!route.api?.bodyFields?.length && <div className="space-y-2">
                                        <h5>Body</h5>
                                        <ul className="list-disc pl-4">
                                            {route.api?.bodyFields?.map((field, i) => (
                                                <li key={field.name + i} className="flex gap-2 items-center">
                                                    <p className="font-medium">{field.jsonName} {field.required &&
                                                        <span className="text-[--red]">*</span>}</p>
                                                    <p className="text-[--muted]">{field.typescriptType}</p>
                                                    {field.descriptions?.map((desc, i) => (
                                                        <p key={desc + i}>{desc}</p>
                                                    ))}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>}

                                    <div className="flex gap-2 items-center">
                                        <p className="font-medium text-[--muted]">Returns</p>
                                        <p className="font-bold text-brand-900">{route.api!.returnTypescriptType}</p>
                                    </div>
                                </AccordionContent>
                            </AccordionItem>
                        ))}
                    </Accordion>

                    <Separator />
                </div>
            ))}
        </div>
    )
}
