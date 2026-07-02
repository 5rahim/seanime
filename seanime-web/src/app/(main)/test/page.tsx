import { PageWrapper } from "@/components/shared/page-wrapper"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Drawer } from "@/components/ui/drawer"
import { DropdownMenu, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"
import { FiEye, FiHeart, FiHelpCircle, FiPlus, FiSearch, FiSettings, FiShare2, FiTrash, FiX } from "react-icons/fi"
import { toast } from "sonner"

const SECTIONS = [
    { id: "buttons", name: "Buttons & Action" },
    { id: "forms", name: "Forms & Controls" },
    { id: "dialogs", name: "Dialogs & Overlays" },
    { id: "feedback", name: "Feedback & Tabs" },
    { id: "layout", name: "Layout & Skeletons" },
]

export default function TestPage() {
    // Global interaction states
    const [globalLoading, setGlobalLoading] = React.useState(false)
    const [globalDisabled, setGlobalDisabled] = React.useState(false)
    const [globalRounded, setGlobalRounded] = React.useState(false)

    // Form states
    const [textValue, setTextValue] = React.useState("Hello Seanime")
    const [passValue, setPassValue] = React.useState("password123")
    const [checkboxVal, setCheckboxVal] = React.useState<boolean | "indeterminate">(true)
    const [switchVal, setSwitchVal] = React.useState(false)
    const [selectVal, setSelectVal] = React.useState("opt2")

    // Overlay states
    const [modal1Open, setModal1Open] = React.useState(false)
    const [modal2Open, setModal2Open] = React.useState(false)
    const [drawerOpen, setDrawerOpen] = React.useState(false)

    // Dismissable badge list
    const [badges, setBadges] = React.useState(["Interactive", "Closable", "Badges"])

    return (
        <PageWrapper className="p-4 md:p-8 space-y-8">
            <div className="relative overflow-hidden rounded-2xl border md:p-8 shadow-md">
                <div className="absolute inset-0 bg-grid-white/[0.02] bg-[size:30px_30px]" />
                <div className="relative z-10 space-y-4">
                    <div className="flex flex-wrap gap-6 items-center bg-black/20 p-4 rounded-xl border border-white/5 w-fit">
                        <Switch
                            label="Simulate Loading"
                            value={globalLoading}
                            onValueChange={setGlobalLoading}
                            size="sm"
                        />
                        <Switch
                            label="Disable Buttons"
                            value={globalDisabled}
                            onValueChange={setGlobalDisabled}
                            size="sm"
                        />
                        <Switch
                            label="Force Rounded"
                            value={globalRounded}
                            onValueChange={setGlobalRounded}
                            size="sm"
                        />
                    </div>
                </div>
            </div>

            {/* Split Page Layout: Sticky Section Nav & Sandbox Sections */}
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
                {/* Side Navigation Anchor Links */}
                <div className="lg:col-span-1 lg:sticky lg:top-20 h-fit space-y-2 bg-gray-900/30 dark:bg-gray-900/40 p-4 rounded-xl border border-white/5">
                    <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2 mb-3">Sections</p>
                    <div className="flex flex-row lg:flex-col overflow-auto gap-1 pb-2 lg:pb-0">
                        {SECTIONS.map((sec) => (
                            <a
                                key={sec.id}
                                href={`#${sec.id}`}
                                className="flex-none whitespace-nowrap px-3 py-2 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-white/5 transition-all"
                            >
                                {sec.name}
                            </a>
                        ))}
                    </div>
                </div>

                {/* Main Content Area */}
                <div className="lg:col-span-3 space-y-12">

                    {/* BUTTONS SECTION */}
                    <section id="buttons" className="space-y-6 scroll-mt-24">
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl md:text-2xl font-bold text-white">Buttons & Action</h2>
                            <Badge intent="gray-solid" size="sm">ui/button</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Button Intents & Variants</CardTitle>
                                <CardDescription>Displaying permutations of intent classes and styled states.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Intents Grid */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Button Intents & Variants</p>
                                    <div className="grid grid-cols-1 sm:grid-cols-[100px_1fr] gap-x-6 gap-y-4 items-center">
                                        <div className="text-xs font-semibold text-gray-500 uppercase">Primary</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="primary-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="primary-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="primary-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="primary-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Success</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="success"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="success-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="success-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="success-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="success-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Warning</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="warning"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="warning-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="warning-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="warning-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="warning-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Alert</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="alert"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="alert-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="alert-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="alert-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Gray</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="gray"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="gray-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="gray-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="gray-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="gray-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">White</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="white"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Solid</Button>
                                            <Button
                                                intent="white-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Outline</Button>
                                            <Button
                                                intent="white-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Subtle</Button>
                                            <Button
                                                intent="white-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Basic</Button>
                                            <Button
                                                intent="white-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Link</Button>
                                        </div>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Aligned Sizes Row */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Aligned Sizes & Icon Button
                                                                                                                Alignment</p>
                                    <div className="space-y-4">
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">XS</span>
                                            <Button
                                                size="xs"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >XS Button</Button>
                                            <Button
                                                size="xs"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >With Icon</Button>
                                            <IconButton
                                                size="xs"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Settings"
                                            />
                                            <IconButton
                                                size="xs"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Favorite"
                                            />
                                        </div>
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">SM</span>
                                            <Button
                                                size="sm"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >SM Button</Button>
                                            <Button
                                                size="sm"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >With Icon</Button>
                                            <IconButton
                                                size="sm"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Settings"
                                            />
                                            <IconButton
                                                size="sm"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Favorite"
                                            />
                                        </div>
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">MD</span>
                                            <Button
                                                size="md"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >MD Button</Button>
                                            <Button
                                                size="md"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >With Icon</Button>
                                            <IconButton
                                                size="md"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Settings"
                                            />
                                            <IconButton
                                                size="md"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Favorite"
                                            />
                                        </div>
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">LG</span>
                                            <Button
                                                size="lg"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >LG Button</Button>
                                            <Button
                                                size="lg"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >With Icon</Button>
                                            <IconButton
                                                size="lg"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Settings"
                                            />
                                            <IconButton
                                                size="lg"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalRounded}
                                                aria-label="Favorite"
                                            />
                                        </div>
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">XL</span>
                                            <Button
                                                size="xl"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >XL Button</Button>
                                            <Button
                                                size="xl"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >With Icon</Button>
                                            <IconButton
                                                size="xl"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Settings"
                                            />
                                            <IconButton
                                                size="xl"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Favorite"
                                            />
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </section>

                    {/* FORMS SECTION */}
                    <section id="forms" className="space-y-6 scroll-mt-24">
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl md:text-2xl font-bold text-white">Forms & Controls</h2>
                            <Badge intent="gray-solid" size="sm">ui/input & selectors</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Form Inputs & Selectors</CardTitle>
                                <CardDescription>Displaying native form components and validations.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Text Inputs */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <TextInput
                                        label="Standard Text Field"
                                        placeholder="Type something..."
                                        value={textValue}
                                        onValueChange={setTextValue}
                                        help="Dynamic text sync is fully active."
                                    />
                                    <TextInput
                                        label="Password Field"
                                        type="password"
                                        placeholder="Enter password..."
                                        value={passValue}
                                        onValueChange={setPassValue}
                                        help="Includes built-in interactive show/hide toggle."
                                    />
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <TextInput
                                        label="Input with Left & Right Icons"
                                        placeholder="Search directories..."
                                        leftIcon={<FiSearch />}
                                        rightIcon={<FiX className="cursor-pointer" onClick={() => setTextValue("")} />}
                                        value={textValue}
                                        onValueChange={setTextValue}
                                    />
                                    <Select
                                        label="Option Selector"
                                        value={selectVal}
                                        onValueChange={setSelectVal}
                                        options={[
                                            { value: "opt1", label: "Option One (Standard)" },
                                            { value: "opt2", label: "Option Two (Recommended)" },
                                            { value: "opt3", label: "Option Three (Disabled)", disabled: true },
                                        ]}
                                        help="Select options from clean overlay menu."
                                    />
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Checkbox & Switch controls */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div className="space-y-4">
                                        <p className="text-sm font-semibold text-gray-300">Checkbox States</p>
                                        <Checkbox
                                            label="Controlled Checkbox"
                                            value={checkboxVal === true}
                                            onValueChange={(checked) => setCheckboxVal(checked)}
                                            help="Supports checked, unchecked, and custom state toggle."
                                        />
                                        <Checkbox
                                            label="Indeterminate State"
                                            value="indeterminate"
                                            help="Useful for parent/child hierarchies."
                                        />
                                        <Checkbox
                                            label="Disabled Option"
                                            disabled
                                            value={true}
                                        />
                                    </div>

                                    <div className="space-y-4">
                                        <p className="text-sm font-semibold text-gray-300">Switch / Toggle States</p>
                                        <Switch
                                            label="Switch Toggle Control"
                                            value={switchVal}
                                            onValueChange={setSwitchVal}
                                            help="Simple binary state switch."
                                        />
                                        <Switch
                                            label="Switch with Tooltip/Help Icon"
                                            value={switchVal}
                                            onValueChange={setSwitchVal}
                                            moreHelp="Additional detailed support description inside this popover."
                                        />
                                        <Switch
                                            label="Switch Align Right"
                                            value={switchVal}
                                            onValueChange={setSwitchVal}
                                            side="right"
                                        />
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </section>

                    {/* DIALOGS SECTION */}
                    <section id="dialogs" className="space-y-6 scroll-mt-24">
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl md:text-2xl font-bold text-white">Dialogs & Overlays</h2>
                            <Badge intent="gray-solid" size="sm">ui/modal & overlays</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Modals, Drawers, Popovers & Tooltips</CardTitle>
                                <CardDescription>Triggers and anchors for layered overlay panels.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                <div className="flex flex-wrap gap-4">
                                    {/* Modal Variant 1: Controlled */}
                                    <Button intent="primary-subtle" leftIcon={<FiShare2 />} onClick={() => setModal1Open(true)}>
                                        Open Modal (Controlled)
                                    </Button>

                                    <Modal
                                        open={modal1Open}
                                        onOpenChange={setModal1Open}
                                        title="System Administration Warning"
                                        description="You are entering the master database control interface."
                                        footer={
                                            <div className="flex gap-2 justify-end w-full">
                                                <Button intent="gray-outline" onClick={() => setModal1Open(false)}>Cancel</Button>
                                                <Button intent="warning" onClick={() => setModal1Open(false)}>Confirm Access</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-4 space-y-3">
                                            <Alert
                                                intent="warning"
                                                title="Critical Node"
                                                description="Ensure you have backups before completing changes."
                                            />
                                            <p className="text-sm text-gray-300">
                                                This database operations action cannot be undone. Changing server states could cause active streams to
                                                terminate abruptly.
                                            </p>
                                        </div>
                                    </Modal>

                                    {/* Modal Variant 2: Uncontrolled with Trigger prop */}
                                    <Modal
                                        trigger={<Button intent="primary-outline">Open Modal (Trigger Prop)</Button>}
                                        title="Standard Media Options"
                                        description="Update settings for active directory folder."
                                        footer={
                                            <div className="flex justify-end gap-2 w-full">
                                                <Button intent="gray" size="sm">Close</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-4 space-y-4">
                                            <TextInput label="Directory Alias" placeholder="Anime collection..." />
                                            <Select
                                                label="Import Speed Limit"
                                                options={[
                                                    { value: "1", label: "Slow (1 MB/s)" },
                                                    { value: "2", label: "Medium (5 MB/s)" },
                                                    { value: "3", label: "Uncapped" },
                                                ]}
                                            />
                                        </div>
                                    </Modal>

                                    {/* Drawer Component */}
                                    <Button intent="success-subtle" leftIcon={<FiPlus />} onClick={() => setDrawerOpen(true)}>
                                        Open Drawer
                                    </Button>

                                    <Drawer
                                        open={drawerOpen}
                                        onOpenChange={setDrawerOpen}
                                        title="Manga Chapter Reader Settings"
                                        description="Customize viewing preferences for local reader canvas."
                                        footer={
                                            <div className="flex gap-2 w-full justify-end">
                                                <Button intent="gray-outline" onClick={() => setDrawerOpen(false)}>Cancel</Button>
                                                <Button intent="success" onClick={() => setDrawerOpen(false)}>Apply Custom Settings</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-6 space-y-6">
                                            <Switch label="Double Page Layout" defaultValue={true} />
                                            <Switch label="Fit to Screen Height" defaultValue={false} />
                                            <Select
                                                label="Transition Animation Style"
                                                options={[
                                                    { value: "fade", label: "Cross-fade Blend" },
                                                    { value: "slide", label: "Horizontal Slide" },
                                                    { value: "none", label: "No Animation" },
                                                ]}
                                            />
                                        </div>
                                    </Drawer>

                                    {/* Popover Component */}
                                    <Popover
                                        trigger={<Button intent="gray-outline" leftIcon={<FiHelpCircle />}>Open Popover</Button>}
                                    >
                                        <div className="space-y-2">
                                            <p className="font-semibold text-white">Need Support?</p>
                                            <p className="text-sm text-gray-300">
                                                Popovers are perfect for showing light contextual fields or details.
                                            </p>
                                            <Button size="xs" intent="primary-subtle" className="w-full">Get Help</Button>
                                        </div>
                                    </Popover>

                                    {/* Dropdown Menu Component */}
                                    <DropdownMenu
                                        trigger={
                                            <Button intent="gray-outline" leftIcon={<FiSettings />}>
                                                Open Dropdown
                                            </Button>
                                        }
                                    >
                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                        <DropdownMenuItem>
                                            <FiEye />
                                            <span>View details</span>
                                        </DropdownMenuItem>
                                        <DropdownMenuItem>
                                            <FiSettings />
                                            <span>Configure</span>
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuGroup>
                                            <DropdownMenuItem className="text-red-300">
                                                <FiTrash />
                                                <span>Delete entry</span>
                                            </DropdownMenuItem>
                                        </DropdownMenuGroup>
                                    </DropdownMenu>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Tooltips */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Hover Tooltips</p>
                                    <div className="flex flex-wrap gap-4">
                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Tooltip Top</Button>}
                                            side="top"
                                        >
                                            Top position hint
                                        </Tooltip>

                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Tooltip Right</Button>}
                                            side="right"
                                        >
                                            Right position hint
                                        </Tooltip>

                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Tooltip Bottom</Button>}
                                            side="bottom"
                                        >
                                            Bottom position hint
                                        </Tooltip>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </section>

                    {/* FEEDBACK & TABS SECTION */}
                    <section id="feedback" className="space-y-6 scroll-mt-24">
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl md:text-2xl font-bold text-white">Feedback & Tabs</h2>
                            <Badge intent="gray-solid" size="sm">ui/alert & ui/badge</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Status Banners & Content Tabs</CardTitle>
                                <CardDescription>Alert feeds, responsive tags, and sub-panels.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Alert Banners */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Alert Banner Layouts</p>
                                    <Alert
                                        intent="info"
                                        title="Informational Alert"
                                        description="This notification informs you that an automated import scan is schedule for midnight."
                                    />
                                    <Alert
                                        intent="success"
                                        title="Successfully Connected"
                                        description="Local server is actively sync'd with AniList database servers."
                                    />
                                    <Alert
                                        intent="warning"
                                        title="Caution Advised"
                                        description="You have 3 unregistered folders which might contain misnamed episodes."
                                    />
                                    <Alert
                                        intent="alert"
                                        title="Connection Dropped"
                                        description="Server failed to authenticate with the API endpoint. Retrying..."
                                        isClosable
                                        onClose={() => console.log("Closed alert")}
                                    />
                                    <Alert
                                        intent="info-basic"
                                        title="Informational Alert"
                                        description="This notification informs you that an automated import scan is schedule for midnight."
                                    />
                                    <Alert
                                        intent="success-basic"
                                        title="Successfully Connected"
                                        description="Local server is actively sync'd with AniList database servers."
                                    />
                                    <Alert
                                        intent="warning-basic"
                                        title="Caution Advised"
                                        description="You have 3 unregistered folders which might contain misnamed episodes."
                                    />
                                    <Alert
                                        intent="alert-basic"
                                        title="Connection Dropped"
                                        description="Server failed to authenticate with the API endpoint. Retrying..."
                                        isClosable
                                        onClose={() => console.log("Closed alert")}
                                    />
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Toast Notifications */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Toast Notifications (Sonner)</p>
                                    <div className="flex flex-wrap gap-3">
                                        <Button
                                            intent="primary-basic"
                                            onClick={() => toast("Default Toast notification")}
                                        >
                                            Default
                                        </Button>
                                        <Button
                                            intent="primary-subtle"
                                            onClick={() => toast.info("Info Toast notification", {
                                                description: "This is a detailed description of the info toast.",
                                            })}
                                        >
                                            Info
                                        </Button>
                                        <Button
                                            intent="success"
                                            onClick={() => toast.success("Successfully sync'd!", {})}
                                        >
                                            Success
                                        </Button>
                                        <Button
                                            intent="warning"
                                            onClick={() => toast.warning("Slow network detected", {
                                                description: "Retrying to fetch the metadata provider.",
                                            })}
                                        >
                                            Warning
                                        </Button>
                                        <Button
                                            intent="alert"
                                            onClick={() => toast.error("Failed to connect", {})}
                                        >
                                            Error
                                        </Button>
                                        <Button
                                            intent="gray-outline"
                                            onClick={() => toast("Undo action toast", {
                                                action: {
                                                    label: "Undo",
                                                    onClick: () => toast("Action undone"),
                                                },
                                            })}
                                        >
                                            With Action
                                        </Button>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Badges */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Badges & Tags</p>

                                    {/* Badge Intents */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Standard Soft Badges</p>
                                        <div className="flex flex-wrap gap-2">
                                            <Badge intent="gray">GraySoft</Badge>
                                            <Badge intent="primary">PrimarySoft</Badge>
                                            <Badge intent="success">SuccessSoft</Badge>
                                            <Badge intent="warning">WarningSoft</Badge>
                                            <Badge intent="alert">AlertSoft</Badge>
                                            <Badge intent="blue">BlueSoft</Badge>
                                            <Badge intent="indigo">IndigoSoft</Badge>
                                            <Badge intent="info">InfoSoft</Badge>
                                            <Badge intent="white">WhiteSoft</Badge>
                                        </div>
                                    </div>

                                    {/* Solid Badges */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Solid / Filled Badges</p>
                                        <div className="flex flex-wrap gap-2">
                                            <Badge intent="primary-solid">PrimarySolid</Badge>
                                            <Badge intent="success-solid">SuccessSolid</Badge>
                                            <Badge intent="warning-solid">WarningSolid</Badge>
                                            <Badge intent="alert-solid">AlertSolid</Badge>
                                            <Badge intent="info-solid">InfoSolid</Badge>
                                            <Badge intent="gray-solid">GraySolid</Badge>
                                        </div>
                                    </div>

                                    {/* Sizes & Interactive */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Badge Sizes & Closable (Interactive)</p>
                                        <div className="flex flex-wrap items-center gap-3">
                                            <Badge size="sm" intent="primary">Small Badge</Badge>
                                            <Badge size="md" intent="primary">Medium Badge</Badge>
                                            <Badge size="lg" intent="primary">Large Badge</Badge>
                                            <Badge size="xl" intent="primary">Extra Large</Badge>
                                            <Separator orientation="vertical" className="h-6 bg-white/10" />
                                            {badges.map(text => (
                                                <Badge
                                                    key={text}
                                                    isClosable
                                                    onClose={() => setBadges(prev => prev.filter(b => b !== text))}
                                                    intent="warning"
                                                >
                                                    {text}
                                                </Badge>
                                            ))}
                                            {badges.length === 0 && (
                                                <Button
                                                    size="xs"
                                                    intent="gray-outline"
                                                    onClick={() => setBadges(["Interactive", "Closable", "Badges"])}
                                                >Reset Tags</Button>
                                            )}
                                        </div>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Tabs */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Tabbed Panels</p>
                                    <Tabs defaultValue="tab1" className="w-full border rounded-xl overflow-hidden bg-black/10">
                                        <TabsList className="bg-gray-900/50 border-b p-0 flex justify-start">
                                            <TabsTrigger value="tab1">Active Queue</TabsTrigger>
                                            <TabsTrigger value="tab2">Download Settings</TabsTrigger>
                                            <TabsTrigger value="tab3">History Logs</TabsTrigger>
                                        </TabsList>
                                        <div className="p-4 min-h-[100px] text-sm text-gray-300">
                                            <TabsContent value="tab1" className="space-y-2">
                                                <p className="font-semibold text-white">Download Queue Status</p>
                                                <p>All items in the scheduler are currently processing. Estimated finish: 14 mins.</p>
                                            </TabsContent>
                                            <TabsContent value="tab2" className="space-y-4">
                                                <p className="font-semibold text-white">Client Overrides</p>
                                                <TextInput label="Target Folder" placeholder="/volume1/media/anime" size="sm" className="max-w-md" />
                                            </TabsContent>
                                            <TabsContent value="tab3">
                                                <p className="font-semibold text-white">Log entries (last 24 hours)</p>
                                                <p className="font-mono text-xs text-brand-400 mt-2 bg-black/40 p-2 rounded">
                                                    [INFO] 2026-07-02 07:11:00 - Scheduled library scan complete. found 0 changes.
                                                </p>
                                            </TabsContent>
                                        </div>
                                    </Tabs>
                                </div>
                            </CardContent>
                        </Card>
                    </section>

                    {/* LAYOUT & SKELETONS SECTION */}
                    <section id="layout" className="space-y-6 scroll-mt-24">
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl md:text-2xl font-bold text-white">Layout & Skeletons</h2>
                            <Badge intent="gray-solid" size="sm">ui/accordion & skeleton</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Accordions & Visual Placeholders</CardTitle>
                                <CardDescription>Folding structures and content loaders.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Accordion */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Accordion Expanders</p>
                                    <Accordion type="single" collapsible className="border rounded-xl bg-black/10 overflow-hidden divide-y">
                                        <AccordionItem value="item-1">
                                            <AccordionTrigger>What is Seanime server sidecar mode?</AccordionTrigger>
                                            <AccordionContent>
                                                Sidecar mode enables Seanime to run locally on desktop computers while executing requests
                                                asynchronously in the background. It integrates directly with MPV media players.
                                            </AccordionContent>
                                        </AccordionItem>
                                        <AccordionItem value="item-2">
                                            <AccordionTrigger>How do I change Torrent streaming caching limits?</AccordionTrigger>
                                            <AccordionContent>
                                                Navigate to Settings &gt; Torrent stream settings. You can set the RAM buffer caching parameters and
                                                maximum download speeds from that dashboard.
                                            </AccordionContent>
                                        </AccordionItem>
                                    </Accordion>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Card Showcase */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Card Component Anatomy</p>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <Card>
                                            <CardHeader>
                                                <CardTitle className="text-base font-bold">Standard Card Title</CardTitle>
                                                <CardDescription>Description or subtitle goes here</CardDescription>
                                            </CardHeader>
                                            <CardContent className="text-sm text-gray-300">
                                                This is the main card body content. You can place standard list items, options, or details in this
                                                area.
                                            </CardContent>
                                            <CardFooter className="justify-between border-t border-white/5 pt-3">
                                                <span className="text-xs text-[--muted]">Last updated: 2 mins ago</span>
                                                <Button size="xs" intent="primary-subtle">Action</Button>
                                            </CardFooter>
                                        </Card>

                                        <Card className="bg-gray-900/40 border-brand-500/20">
                                            <CardHeader>
                                                <CardTitle className="text-base font-bold text-brand-300">Styled Banner Card</CardTitle>
                                                <CardDescription className="text-brand-400/80">Highlight critical system details</CardDescription>
                                            </CardHeader>
                                            <CardContent className="text-sm text-gray-200">
                                                Premium card styling using subtle background colors and higher contrast borders to separate sections.
                                            </CardContent>
                                            <CardFooter className="justify-end gap-2">
                                                <Button size="xs" intent="gray-basic">Ignore</Button>
                                                <Button size="xs" intent="primary">Acknowledge</Button>
                                            </CardFooter>
                                        </Card>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Skeletons */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Skeleton Loaders</p>
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                                        {/* Skeleton Example 1 */}
                                        <div className="space-y-3 border p-4 rounded-xl bg-black/5">
                                            <Skeleton className="h-6 w-1/3" />
                                            <Skeleton className="h-24 w-full" />
                                            <div className="flex gap-2">
                                                <Skeleton className="h-8 w-16" />
                                                <Skeleton className="h-8 w-16" />
                                            </div>
                                        </div>

                                        {/* Skeleton Example 2 */}
                                        <div className="flex items-center gap-3 border p-4 rounded-xl bg-black/5">
                                            <Skeleton className="h-12 w-12 rounded-full shrink-0" />
                                            <div className="space-y-2 w-full">
                                                <Skeleton className="h-4 w-3/4" />
                                                <Skeleton className="h-3 w-1/2" />
                                            </div>
                                        </div>

                                        {/* Skeleton Example 3 */}
                                        <div className="space-y-3 border p-4 rounded-xl bg-black/5">
                                            <div className="flex justify-between items-center">
                                                <Skeleton className="h-4 w-1/4" />
                                                <Skeleton className="h-4 w-12 rounded-full" />
                                            </div>
                                            <Skeleton className="h-10 w-full" />
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </section>
                </div>
            </div>
        </PageWrapper>
    )
}
