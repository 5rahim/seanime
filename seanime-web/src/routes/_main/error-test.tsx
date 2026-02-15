import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/error-test")({
    component: ErrorTest,
})

function ErrorTest() {
    throw new Error("This is a test error")

    return (
        <div className="p-4">
            <h1>Error Test</h1>
        </div>
    )
}
