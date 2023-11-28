import React from "react"

/* -------------------------------------------------------------------------------------------------
 * ShowOnly
 * -----------------------------------------------------------------------------------------------*/

interface ShowOnlyProps {
    children?: React.ReactNode
    when: boolean | undefined
}

export const ShowOnly: React.FC<ShowOnlyProps> = (props) => {

    const { children, when } = props

    return (
        <>
            {when ? children : null}
        </>
    )

}

ShowOnly.displayName = "ShowOnly"
