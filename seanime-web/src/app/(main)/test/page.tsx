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
    const [textValue, setTextValue] = React.useState("Lorem ipsum")
    const [passValue, setPassValue] = React.useState("lorem123")
    const [checkboxVal, setCheckboxVal] = React.useState<boolean | "indeterminate">(true)
    const [switchVal, setSwitchVal] = React.useState(false)
    const [selectVal, setSelectVal] = React.useState("opt2")

    // Overlay states
    const [modal1Open, setModal1Open] = React.useState(false)
    const [modal2Open, setModal2Open] = React.useState(false)
    const [drawerOpen, setDrawerOpen] = React.useState(false)

    // Dismissable badge list
    const [badges, setBadges] = React.useState(["Lorem", "Ipsum", "Dolor"])

    return (
        <PageWrapper className="p-4 md:p-8 space-y-8">
            <div className="relative overflow-hidden rounded-2xl border md:p-8 shadow-md">
                <div className="absolute inset-0 bg-grid-white/[0.02] bg-[size:30px_30px]" />
                <div className="relative z-10 space-y-4">
                    <div className="flex flex-wrap gap-6 items-center bg-black/20 p-4 rounded-xl border border-white/5 w-fit">
                        <Switch
                            label="Loading"
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
                            label="Rounded"
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
                            <Badge intent="gray-solid" size="sm">Lorem</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Lorem Ipsum</CardTitle>
                                <CardDescription>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Intents Grid */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <div className="grid grid-cols-1 sm:grid-cols-[100px_1fr] gap-x-6 gap-y-4 items-center">
                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="primary-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="primary-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="primary-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="primary-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="success"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="success-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="success-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="success-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="success-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="warning"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="warning-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="warning-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="warning-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="warning-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="alert"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="alert-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="alert-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="alert-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="gray"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="gray-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="gray-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="gray-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="gray-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>

                                        <div className="text-xs font-semibold text-gray-500 uppercase">Lorem</div>
                                        <div className="flex flex-wrap gap-2.5">
                                            <Button
                                                intent="white"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                intent="white-outline"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Ipsum</Button>
                                            <Button
                                                intent="white-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Dolor</Button>
                                            <Button
                                                intent="white-basic"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Sit</Button>
                                            <Button
                                                intent="white-link"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Amet</Button>
                                        </div>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Aligned Sizes Row */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <div className="space-y-4">
                                        <div className="flex flex-wrap items-center gap-4">
                                            <span className="text-xs font-mono text-gray-500 w-8">XS</span>
                                            <Button
                                                size="xs"
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <Button
                                                size="xs"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <IconButton
                                                size="xs"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
                                            />
                                            <IconButton
                                                size="xs"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
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
                                            >Lorem</Button>
                                            <Button
                                                size="sm"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <IconButton
                                                size="sm"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
                                            />
                                            <IconButton
                                                size="sm"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
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
                                            >Lorem</Button>
                                            <Button
                                                size="md"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <IconButton
                                                size="md"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
                                            />
                                            <IconButton
                                                size="md"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
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
                                            >Lorem</Button>
                                            <Button
                                                size="lg"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <IconButton
                                                size="lg"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
                                            />
                                            <IconButton
                                                size="lg"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalRounded}
                                                aria-label="Lorem"
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
                                            >Lorem</Button>
                                            <Button
                                                size="xl"
                                                intent="gray-outline"
                                                leftIcon={<FiPlus />}
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                            >Lorem</Button>
                                            <IconButton
                                                size="xl"
                                                icon={<FiSettings />}
                                                intent="primary"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
                                            />
                                            <IconButton
                                                size="xl"
                                                icon={<FiHeart />}
                                                intent="alert-subtle"
                                                loading={globalLoading}
                                                disabled={globalDisabled}
                                                rounded={globalRounded}
                                                aria-label="Lorem"
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
                            <Badge intent="gray-solid" size="sm">Lorem</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Lorem Ipsum</CardTitle>
                                <CardDescription>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Text Inputs */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <TextInput
                                        label="Lorem Ipsum"
                                        placeholder="Lorem ipsum..."
                                        value={textValue}
                                        onValueChange={setTextValue}
                                        help="Lorem ipsum dolor sit amet."
                                    />
                                    <TextInput
                                        label="Lorem Ipsum"
                                        type="password"
                                        placeholder="Lorem ipsum..."
                                        value={passValue}
                                        onValueChange={setPassValue}
                                        help="Lorem ipsum dolor sit amet."
                                    />
                                </div>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <TextInput
                                        label="Lorem Ipsum"
                                        placeholder="Lorem ipsum..."
                                        leftIcon={<FiSearch />}
                                        rightIcon={<FiX className="cursor-pointer" onClick={() => setTextValue("")} />}
                                        value={textValue}
                                        onValueChange={setTextValue}
                                    />
                                    <Select
                                        label="Lorem Ipsum"
                                        value={selectVal}
                                        onValueChange={setSelectVal}
                                        options={[
                                            { value: "opt1", label: "Lorem Ipsum" },
                                            { value: "opt2", label: "Lorem Ipsum" },
                                            { value: "opt3", label: "Lorem Ipsum", disabled: true },
                                        ]}
                                        help="Lorem ipsum dolor sit amet."
                                    />
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Checkbox & Switch controls */}
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div className="space-y-4">
                                        <p className="text-sm font-semibold text-gray-300">Lorem Ipsum</p>
                                        <Checkbox
                                            label="Lorem Ipsum"
                                            value={checkboxVal === true}
                                            onValueChange={(checked) => setCheckboxVal(checked)}
                                            help="Lorem ipsum dolor sit amet."
                                        />
                                        <Checkbox
                                            label="Lorem Ipsum"
                                            value="indeterminate"
                                            help="Lorem ipsum dolor sit amet."
                                        />
                                        <Checkbox
                                            label="Lorem Ipsum"
                                            disabled
                                            value={true}
                                        />
                                    </div>

                                    <div className="space-y-4">
                                        <p className="text-sm font-semibold text-gray-300">Lorem Ipsum</p>
                                        <Switch
                                            label="Lorem Ipsum"
                                            value={switchVal}
                                            onValueChange={setSwitchVal}
                                            help="Lorem ipsum dolor sit amet."
                                        />
                                        <Switch
                                            label="Lorem Ipsum"
                                            value={switchVal}
                                            onValueChange={setSwitchVal}
                                            moreHelp="Lorem ipsum dolor sit amet."
                                        />
                                        <Switch
                                            label="Lorem Ipsum"
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
                            <Badge intent="gray-solid" size="sm">Lorem</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Lorem Ipsum</CardTitle>
                                <CardDescription>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-6">
                                <div className="flex flex-wrap gap-4">
                                    {/* Modal Variant 1: Controlled */}
                                    <Button intent="primary-subtle" leftIcon={<FiShare2 />} onClick={() => setModal1Open(true)}>
                                        Lorem Ipsum
                                    </Button>

                                    <Modal
                                        open={modal1Open}
                                        onOpenChange={setModal1Open}
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                        footer={
                                            <div className="flex gap-2 justify-end w-full">
                                                <Button intent="gray-outline" onClick={() => setModal1Open(false)}>Lorem</Button>
                                                <Button intent="warning" onClick={() => setModal1Open(false)}>Lorem</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-4 space-y-3">
                                            <Alert
                                                intent="warning"
                                                title="Lorem Ipsum"
                                                description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                            />
                                            <p className="text-sm text-gray-300">
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
                                                dolore magna aliqua.
                                            </p>
                                        </div>
                                    </Modal>

                                    {/* Modal Variant 2: Uncontrolled with Trigger prop */}
                                    <Modal
                                        trigger={<Button intent="primary-outline">Lorem Ipsum</Button>}
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet."
                                        footer={
                                            <div className="flex justify-end gap-2 w-full">
                                                <Button intent="gray" size="sm">Lorem</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-4 space-y-4">
                                            <TextInput label="Lorem Ipsum" placeholder="Lorem ipsum..." />
                                            <Select
                                                label="Lorem Ipsum"
                                                options={[
                                                    { value: "1", label: "Lorem" },
                                                    { value: "2", label: "Ipsum" },
                                                    { value: "3", label: "Dolor" },
                                                ]}
                                            />
                                        </div>
                                    </Modal>

                                    {/* Drawer Component */}
                                    <Button intent="success-subtle" leftIcon={<FiPlus />} onClick={() => setDrawerOpen(true)}>
                                        Lorem Ipsum
                                    </Button>

                                    <Drawer
                                        open={drawerOpen}
                                        onOpenChange={setDrawerOpen}
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet."
                                        footer={
                                            <div className="flex gap-2 w-full justify-end">
                                                <Button intent="gray-outline" onClick={() => setDrawerOpen(false)}>Lorem</Button>
                                                <Button intent="success" onClick={() => setDrawerOpen(false)}>Lorem</Button>
                                            </div>
                                        }
                                    >
                                        <div className="py-6 space-y-6">
                                            <Switch label="Lorem Ipsum" defaultValue={true} />
                                            <Switch label="Lorem Ipsum" defaultValue={false} />
                                            <Select
                                                label="Lorem Ipsum"
                                                options={[
                                                    { value: "fade", label: "Lorem" },
                                                    { value: "slide", label: "Ipsum" },
                                                    { value: "none", label: "Dolor" },
                                                ]}
                                            />
                                        </div>
                                    </Drawer>

                                    {/* Popover Component */}
                                    <Popover
                                        trigger={<Button intent="gray-outline" leftIcon={<FiHelpCircle />}>Lorem Ipsum</Button>}
                                    >
                                        <div className="space-y-2">
                                            <p className="font-semibold text-white">Lorem Ipsum</p>
                                            <p className="text-sm text-gray-300">
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit.
                                            </p>
                                            <Button size="xs" intent="primary-subtle" className="w-full">Lorem</Button>
                                        </div>
                                    </Popover>

                                    {/* Dropdown Menu Component */}
                                    <DropdownMenu
                                        trigger={
                                            <Button intent="gray-outline" leftIcon={<FiSettings />}>
                                                Lorem Ipsum
                                            </Button>
                                        }
                                    >
                                        <DropdownMenuLabel>Lorem Ipsum</DropdownMenuLabel>
                                        <DropdownMenuItem>
                                            <FiEye />
                                            <span>Lorem Ipsum</span>
                                        </DropdownMenuItem>
                                        <DropdownMenuItem>
                                            <FiSettings />
                                            <span>Lorem Ipsum</span>
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuGroup>
                                            <DropdownMenuItem className="text-red-300">
                                                <FiTrash />
                                                <span>Lorem Ipsum</span>
                                            </DropdownMenuItem>
                                        </DropdownMenuGroup>
                                    </DropdownMenu>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Tooltips */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <div className="flex flex-wrap gap-4">
                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Lorem Ipsum</Button>}
                                            side="top"
                                        >
                                            Lorem Ipsum
                                        </Tooltip>

                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Lorem Ipsum</Button>}
                                            side="right"
                                        >
                                            Lorem Ipsum
                                        </Tooltip>

                                        <Tooltip
                                            trigger={<Button size="sm" intent="gray-outline">Lorem Ipsum</Button>}
                                            side="bottom"
                                        >
                                            Lorem Ipsum
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
                            <Badge intent="gray-solid" size="sm">Lorem</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Lorem Ipsum</CardTitle>
                                <CardDescription>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Alert Banners */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <Alert
                                        intent="info"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="success"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="warning"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="alert"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                        isClosable
                                        onClose={() => console.log("Closed alert")}
                                    />
                                    <Alert
                                        intent="info-basic"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="success-basic"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="warning-basic"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                    />
                                    <Alert
                                        intent="alert-basic"
                                        title="Lorem Ipsum"
                                        description="Lorem ipsum dolor sit amet, consectetur adipiscing elit."
                                        isClosable
                                        onClose={() => console.log("Closed alert")}
                                    />
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Toast Notifications */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <div className="flex flex-wrap gap-3">
                                        <Button
                                            intent="primary-basic"
                                            onClick={() => toast("Lorem ipsum dolor sit amet")}
                                        >
                                            Lorem
                                        </Button>
                                        <Button
                                            intent="primary-subtle"
                                            onClick={() => toast.info("Lorem ipsum dolor sit amet", {
                                                description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
                                            })}
                                        >
                                            Lorem
                                        </Button>
                                        <Button
                                            intent="success"
                                            onClick={() => toast.success("Lorem ipsum dolor sit amet", {})}
                                        >
                                            Lorem
                                        </Button>
                                        <Button
                                            intent="warning"
                                            onClick={() => toast.warning("Lorem ipsum dolor sit amet", {
                                                description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
                                            })}
                                        >
                                            Lorem
                                        </Button>
                                        <Button
                                            intent="alert"
                                            onClick={() => toast.error("Lorem ipsum dolor sit amet", {})}
                                        >
                                            Lorem
                                        </Button>
                                        <Button
                                            intent="gray-outline"
                                            onClick={() => toast("Lorem ipsum dolor sit amet", {
                                                action: {
                                                    label: "Lorem",
                                                    onClick: () => toast("Lorem ipsum dolor sit amet"),
                                                },
                                            })}
                                        >
                                            Lorem
                                        </Button>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Badges */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>

                                    {/* Badge Intents */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Lorem Ipsum</p>
                                        <div className="flex flex-wrap gap-2">
                                            <Badge intent="gray">Lorem</Badge>
                                            <Badge intent="primary">Lorem</Badge>
                                            <Badge intent="success">Lorem</Badge>
                                            <Badge intent="warning">Lorem</Badge>
                                            <Badge intent="alert">Lorem</Badge>
                                            <Badge intent="blue">Lorem</Badge>
                                            <Badge intent="indigo">Lorem</Badge>
                                            <Badge intent="info">Lorem</Badge>
                                            <Badge intent="white">Lorem</Badge>
                                        </div>
                                    </div>

                                    {/* Solid Badges */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Lorem Ipsum</p>
                                        <div className="flex flex-wrap gap-2">
                                            <Badge intent="primary-solid">Lorem</Badge>
                                            <Badge intent="success-solid">Lorem</Badge>
                                            <Badge intent="warning-solid">Lorem</Badge>
                                            <Badge intent="alert-solid">Lorem</Badge>
                                            <Badge intent="info-solid">Lorem</Badge>
                                            <Badge intent="gray-solid">Lorem</Badge>
                                        </div>
                                    </div>

                                    {/* Sizes & Interactive */}
                                    <div className="space-y-2">
                                        <p className="text-xs text-gray-500">Lorem Ipsum</p>
                                        <div className="flex flex-wrap items-center gap-3">
                                            <Badge size="sm" intent="primary">Lorem</Badge>
                                            <Badge size="md" intent="primary">Lorem</Badge>
                                            <Badge size="lg" intent="primary">Lorem</Badge>
                                            <Badge size="xl" intent="primary">Lorem</Badge>
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
                                                    onClick={() => setBadges(["Lorem", "Ipsum", "Dolor"])}
                                                >Lorem</Button>
                                            )}
                                        </div>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Tabs */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <Tabs defaultValue="tab1" className="w-full border rounded-xl overflow-hidden bg-black/10">
                                        <TabsList className="bg-gray-900/50 border-b p-0 flex justify-start">
                                            <TabsTrigger value="tab1">Lorem</TabsTrigger>
                                            <TabsTrigger value="tab2">Ipsum</TabsTrigger>
                                            <TabsTrigger value="tab3">Dolor</TabsTrigger>
                                        </TabsList>
                                        <div className="p-4 min-h-[100px] text-sm text-gray-300">
                                            <TabsContent value="tab1" className="space-y-2">
                                                <p className="font-semibold text-white">Lorem Ipsum</p>
                                                <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
                                            </TabsContent>
                                            <TabsContent value="tab2" className="space-y-4">
                                                <p className="font-semibold text-white">Lorem Ipsum</p>
                                                <TextInput label="Lorem Ipsum" placeholder="Lorem ipsum..." size="sm" className="max-w-md" />
                                            </TabsContent>
                                            <TabsContent value="tab3">
                                                <p className="font-semibold text-white">Lorem Ipsum</p>
                                                <p className="font-mono text-xs text-brand-400 mt-2 bg-black/40 p-2 rounded">
                                                    Lorem ipsum dolor sit amet, consectetur adipiscing elit.
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
                            <Badge intent="gray-solid" size="sm">Lorem</Badge>
                        </div>
                        <Card className="border-white/5 bg-gray-900/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Lorem Ipsum</CardTitle>
                                <CardDescription>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-8">
                                {/* Accordion */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <Accordion type="single" collapsible className="border rounded-xl bg-black/10 overflow-hidden divide-y">
                                        <AccordionItem value="item-1">
                                            <AccordionTrigger>Lorem ipsum dolor sit amet?</AccordionTrigger>
                                            <AccordionContent>
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
                                                dolore magna aliqua.
                                            </AccordionContent>
                                        </AccordionItem>
                                        <AccordionItem value="item-2">
                                            <AccordionTrigger>Lorem ipsum dolor sit amet?</AccordionTrigger>
                                            <AccordionContent>
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
                                                dolore magna aliqua.
                                            </AccordionContent>
                                        </AccordionItem>
                                    </Accordion>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Card Showcase */}
                                <div className="space-y-3">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <Card>
                                            <CardHeader>
                                                <CardTitle className="text-base font-bold">Lorem Ipsum</CardTitle>
                                                <CardDescription>Lorem ipsum dolor sit amet</CardDescription>
                                            </CardHeader>
                                            <CardContent className="text-sm text-gray-300">
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
                                                dolore magna aliqua.
                                            </CardContent>
                                            <CardFooter className="justify-between border-t border-white/5 pt-3">
                                                <span className="text-xs text-[--muted]">Lorem ipsum</span>
                                                <Button size="xs" intent="primary-subtle">Lorem</Button>
                                            </CardFooter>
                                        </Card>

                                        <Card className="bg-gray-900/40 border-brand-500/20">
                                            <CardHeader>
                                                <CardTitle className="text-base font-bold text-brand-300">Lorem Ipsum</CardTitle>
                                                <CardDescription className="text-brand-400/80">Lorem ipsum dolor sit amet</CardDescription>
                                            </CardHeader>
                                            <CardContent className="text-sm text-gray-200">
                                                Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
                                                dolore magna aliqua.
                                            </CardContent>
                                            <CardFooter className="justify-end gap-2">
                                                <Button size="xs" intent="gray-basic">Lorem</Button>
                                                <Button size="xs" intent="primary">Lorem</Button>
                                            </CardFooter>
                                        </Card>
                                    </div>
                                </div>

                                <Separator className="bg-white/5" />

                                {/* Skeletons */}
                                <div className="space-y-4">
                                    <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Lorem Ipsum</p>
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
