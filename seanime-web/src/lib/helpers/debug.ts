import chalk from "chalk"

export const logger = (prefix: string) => {

    return {
        info: (...data: any[]) => {
            console.log(chalk.blueBright(`[${prefix}]`) + ": ", ...data)
        },
        warning: (...data: any[]) => {
            console.log(chalk.yellow(`[${prefix}]` + ": "), ...data)
        },
        success: (...data: any[]) => {
            console.log(chalk.greenBright(`[${prefix}]` + ": "), ...data)
        },
        error: (...data: any[]) => {
            console.log(chalk.redBright(`[${prefix}]` + ": "), ...data)
        },
    }

}
