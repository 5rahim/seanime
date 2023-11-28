"use client"

import React, { useLayoutEffect, useState } from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import { DropdownMenu } from "../dropdown-menu"
import { HexColorPicker } from "react-colorful"
import { TextInput, TextInputProps } from "../text-input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ColorInputAnatomy = defineStyleAnatomy({
    colorInput: cva([
        "UI-ColorInput__root",
        "w-6 h-6 rounded-md -ml-1 border border-[--border]"
    ]),
    colorPickerContainer: cva([
        "UI-ColorInput__colorPickerContainer",
        "flex w-full justify-center p-2"
    ])
})

/* -------------------------------------------------------------------------------------------------
 * ColorInput
 * -----------------------------------------------------------------------------------------------*/

export interface ColorInputProps extends ComponentWithAnatomy<typeof ColorInputAnatomy>,
    Omit<TextInputProps, "onChange" | "value" | "defaultValue"> {
    children?: React.ReactNode
    onChange?: (value: string) => void
    value?: string
    defaultValue?: string
}

export const ColorInput: React.FC<ColorInputProps> = React.forwardRef((props, ref) => {

    const {
        children,
        colorInputClassName,
        colorPickerContainerClassName,
        className,
        value,
        onChange,
        defaultValue = "#5e28c2",
        ...rest
    } = props

    const [color, setColor] = useState(defaultValue ?? value)

    // Control the value
    useLayoutEffect(() => {
        if (value) setColor(value)
    }, [value])

    return (
        <DropdownMenu
            trigger={
                <TextInput
                    value={color}
                    onChange={e => setColor(e.target.value)}
                    leftIcon={
                        <div className={cn(ColorInputAnatomy.colorInput(), colorInputClassName)}
                             style={{ backgroundColor: color }}/>
                    }
                    ref={ref}
                    {...rest}
                />
            }
            menuClassName={"w-full block"}
            dropdownClassName={"right-[inherit] left-0"}
        >
            <div className={cn(ColorInputAnatomy.colorPickerContainer(), colorPickerContainerClassName)}>
                <HexColorPicker
                    color={color}
                    onChange={value => {
                        onChange && onChange(value)
                        setColor(value)
                    }}
                />
            </div>
        </DropdownMenu>
    )

})

ColorInput.displayName = "ColorInput"
