import chalk from "chalk"

export const logger = (prefix: string) => {

    return {
        info: (...data: any[]) => {
            console.log(chalk.blue(`[${prefix}]`) + " ", ...data)
        },
        warning: (...data: any[]) => {
            console.log(chalk.yellow(`[${prefix}]`) + " ", ...data)
        },
        success: (...data: any[]) => {
            console.log(chalk.green(`[${prefix}]`) + " ", ...data)
        },
        error: (...data: any[]) => {
            console.log(chalk.red(`[${prefix}]`) + " ", ...data)
        },
    }

}
