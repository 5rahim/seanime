import { Menu, useAudioOptions } from "@vidstack/react"

export function MediastreamAudioSubmenu() {
    const options = useAudioOptions(),
        hint = options.selectedTrack?.label
    return (
        <Menu.Root>
            <Menu.Button disabled={options.disabled}>Audio ({hint})</Menu.Button>
            <Menu.Content>
                <Menu.RadioGroup value={options.selectedValue}>
                    {options.map(({ label, value, select }) => (
                        <Menu.Radio value={value} onSelect={select} key={value}>
                            {label}
                        </Menu.Radio>
                    ))}
                </Menu.RadioGroup>
            </Menu.Content>
        </Menu.Root>
    )
}
