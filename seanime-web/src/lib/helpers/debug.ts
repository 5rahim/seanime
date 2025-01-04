export const logger = (prefix: string) => {

    return {
        info: (...data: any[]) => {
            console.log(`[${prefix}]` + " ", ...data)
        },
        warning: (...data: any[]) => {
            console.log(`[${prefix}]` + " ", ...data)
        },
        success: (...data: any[]) => {
            console.log(`[${prefix}]` + " ", ...data)
        },
        error: (...data: any[]) => {
            console.log(`[${prefix}]` + " ", ...data)
        },
    }

}
