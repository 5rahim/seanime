include(":core")

File(rootDir, "src").eachDir { dir ->
    dir.eachDir { subdir ->
        val name = ":extensions:individual:${dir.name}:${subdir.name}"
        include(name)
        project(name).projectDir = File("src/${dir.name}/${subdir.name}")
    }
}

fun File.eachDir(block: (File) -> Unit) {
    listFiles()?.filter { it.isDirectory }?.forEach { block(it) }
}
