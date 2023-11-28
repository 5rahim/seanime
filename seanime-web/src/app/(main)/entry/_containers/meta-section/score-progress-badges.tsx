import { Nullish } from "@/types/common"
import { Badge } from "@/components/ui/badge"
import React from "react"
import { BiStar } from "@react-icons/all-files/bi/BiStar"

export function ScoreProgressBadges({ score, progress, episodes }: {
    score: Nullish<number>,
    progress: Nullish<number>,
    episodes: Nullish<number>
}) {

    const scoreColor = score ? (
        score < 5 ? "bg-red-500" :
            score < 7 ? "bg-orange-500" :
                score < 9 ? "bg-green-500" :
                    "bg-brand-500 text-white"
    ) : ""

    return (
        <>
            {!!score && <Badge leftIcon={<BiStar/>} size={"xl"} intent={"primary-solid"} className={scoreColor}>
                {score}
            </Badge>}
            <Badge
                size={"xl"}
                className={"!text-lg font-bold !text-yellow-50"}
            >
                {`${progress ?? 0}/${episodes || "-"}`}
            </Badge>
        </>
    )

}