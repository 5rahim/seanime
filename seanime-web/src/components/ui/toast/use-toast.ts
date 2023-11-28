import toast from "react-hot-toast"

/* -------------------------------------------------------------------------------------------------
 * useToast
 * - You can add more types
 * -----------------------------------------------------------------------------------------------*/

export const useToast = () => {

    return {
        success: (message?: string) => {

            toast.success(message ?? "")

        },
        error: (message?: string) => {

            toast.error(message ?? "")

        },
    }

}
