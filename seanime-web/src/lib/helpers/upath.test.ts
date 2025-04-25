import { describe, expect, it } from "vitest"
import { upath } from "./upath"

describe("upath", () => {
    describe("core functions", () => {
        describe("join", () => {
            it("should join paths correctly", () => {
                expect(upath.join("a", "b", "c")).toBe("a/b/c")
                expect(upath.join("/a", "b", "c")).toBe("/a/b/c")
                expect(upath.join("a", "/b", "c")).toBe("a/b/c")
                expect(upath.join("")).toBe(".")
                expect(upath.join("a", "")).toBe("a")
                expect(upath.join("a/", "b")).toBe("a/b")
                expect(upath.join("a//", "b")).toBe("a/b")
                expect(upath.join("a", "..", "b")).toBe("b")
                expect(upath.join("a/b", "..")).toBe("a")
            })

            it("should handle Windows-style paths", () => {
                expect(upath.join("a\\b", "c")).toBe("a/b/c")
                expect(upath.join("a", "b\\c")).toBe("a/b/c")
                expect(upath.join("C:", "b\\c")).toBe("C:/b/c")
            })
        })

        describe("resolve", () => {
            it("should resolve paths correctly", () => {
                expect(upath.resolve("a", "b", "c")).toBe("/a/b/c")
                expect(upath.resolve("/a", "b", "c")).toBe("/a/b/c")
                expect(upath.resolve("/a", "/b", "c")).toBe("/b/c")
                expect(upath.resolve("a", "..", "b")).toBe("/b")
                expect(upath.resolve("a/b", "..")).toBe("/a")
            })

            it("should handle absolute paths", () => {
                expect(upath.resolve("/a/b", "c")).toBe("/a/b/c")
                expect(upath.resolve("/a/b", "/c")).toBe("/c")
            })
        })

        describe("normalize", () => {
            it("should normalize paths correctly", () => {
                expect(upath.normalize("a/b/c")).toBe("a/b/c")
                expect(upath.normalize("/a/b/c")).toBe("/a/b/c")
                expect(upath.normalize("a//b/c")).toBe("a/b/c")
                expect(upath.normalize("a/b/../c")).toBe("a/c")
                expect(upath.normalize("/a/b/../c")).toBe("/a/c")
                expect(upath.normalize("a/b/./c")).toBe("a/b/c")
                expect(upath.normalize("a/b/c/")).toBe("a/b/c/")
                expect(upath.normalize("a/b/c//")).toBe("a/b/c/")
                expect(upath.normalize("")).toBe(".")
                expect(upath.normalize(".")).toBe(".")
                expect(upath.normalize("..")).toBe("..")
            })

            it("should handle special cases", () => {
                expect(upath.normalize("/")).toBe("/")
                expect(upath.normalize("//")).toBe("/")
                expect(upath.normalize("//server/share")).toBe("/server/share")
                expect(upath.normalize("a/../..")).toBe("..")
                expect(upath.normalize("/a/../..")).toBe("/")
            })
        })

        describe("isAbsolute", () => {
            it("should identify absolute paths", () => {
                expect(upath.isAbsolute("/a/b")).toBe(true)
                expect(upath.isAbsolute("/a")).toBe(true)
                expect(upath.isAbsolute("/")).toBe(true)
                expect(upath.isAbsolute("//server/share")).toBe(true)
                expect(upath.isAbsolute("C:\\a\\b")).toBe(true)
            })

            it("should identify relative paths", () => {
                expect(upath.isAbsolute("a/b")).toBe(false)
                expect(upath.isAbsolute(".")).toBe(false)
                expect(upath.isAbsolute("..")).toBe(false)
                expect(upath.isAbsolute("")).toBe(false)
            })
        })

        describe("dirname", () => {
            it("should get directory path correctly", () => {
                expect(upath.dirname("/a/b/c")).toBe("/a/b")
                expect(upath.dirname("/a/b")).toBe("/a")
                expect(upath.dirname("/a")).toBe("/")
                expect(upath.dirname("a/b")).toBe("a")
                expect(upath.dirname("a")).toBe(".")
                expect(upath.dirname(".")).toBe(".")
                expect(upath.dirname("")).toBe(".")
            })

            it("should handle trailing slashes", () => {
                expect(upath.dirname("/a/b/c/")).toBe("/a/b")
                expect(upath.dirname("a/b/")).toBe("a")
            })
        })

        describe("basename", () => {
            it("should extract basename correctly", () => {
                expect(upath.basename("/a/b/c")).toBe("c")
                expect(upath.basename("/a/b/c/")).toBe("c")
                expect(upath.basename("/a/b")).toBe("b")
                expect(upath.basename("/a")).toBe("a")
                expect(upath.basename("a/b")).toBe("b")
                expect(upath.basename("a")).toBe("a")
                expect(upath.basename(".")).toBe(".")
                expect(upath.basename("")).toBe("")
            })

            it("should handle extensions", () => {
                expect(upath.basename("/a/b/c.txt")).toBe("c.txt")
                expect(upath.basename("/a/b/c.txt", ".txt")).toBe("c")
                expect(upath.basename("a.txt", ".txt")).toBe("a")
                expect(upath.basename(".txt", ".txt")).toBe("")
                expect(upath.basename(".bashrc")).toBe(".bashrc")
                expect(upath.basename(".bashrc", ".bashrc")).toBe("")
            })

            it("should handle trailing slashes", () => {
                expect(upath.basename("/a/b/c/")).toBe("c")
                expect(upath.basename("a/b/")).toBe("b")
            })
        })

        describe("extname", () => {
            it("should extract extension correctly", () => {
                expect(upath.extname("a.txt")).toBe(".txt")
                expect(upath.extname("/a/b/c.txt")).toBe(".txt")
                expect(upath.extname("/a/b.c/d")).toBe("")
                expect(upath.extname("a")).toBe("")
                expect(upath.extname("a.")).toBe(".")
                expect(upath.extname(".txt")).toBe("")
                expect(upath.extname("a.b.c")).toBe(".c")
                expect(upath.extname("")).toBe("")
            })
        })

        describe("format", () => {
            it("should format path objects correctly", () => {
                expect(upath.format({ root: "/", dir: "/a/b", base: "c.txt" })).toBe("/a/b/c.txt")
                expect(upath.format({ dir: "a/b", base: "c.txt" })).toBe("a/b/c.txt")
                expect(upath.format({ root: "/", base: "c.txt" })).toBe("/c.txt")
                expect(upath.format({ root: "//" })).toBe("//")
                expect(upath.format({ dir: "a/b" })).toBe("a/b")
                expect(upath.format({ name: "file", ext: ".txt" })).toBe("file.txt")
            })
        })

        describe("parse", () => {
            it("should parse paths correctly", () => {
                expect(upath.parse("/a/b/c.txt")).toEqual({
                    root: "/",
                    dir: "/a/b",
                    base: "c.txt",
                    ext: ".txt",
                    name: "c",
                })
                expect(upath.parse("a/b/c")).toEqual({
                    root: "",
                    dir: "a/b",
                    base: "c",
                    ext: "",
                    name: "c",
                })
                expect(upath.parse(".bashrc")).toEqual({
                    root: "",
                    dir: "",
                    base: ".bashrc",
                    ext: "",
                    name: ".bashrc",
                })
                expect(upath.parse("//server/share/file.txt")).toEqual({
                    root: "//",
                    dir: "//server/share",
                    base: "file.txt",
                    ext: ".txt",
                    name: "file",
                })
            })
        })

        describe("relative", () => {
            it("should calculate relative paths correctly", () => {
                expect(upath.relative("/a/b/c", "/a/b/d")).toBe("../d")
                expect(upath.relative("/a/b", "/a/c")).toBe("../c")
                expect(upath.relative("/a/b", "/a/b/c")).toBe("c")
                expect(upath.relative("/a/b/c", "/a/b")).toBe("..")
                expect(upath.relative("/a/b/c", "/d/e/f")).toBe("../../../d/e/f")
                expect(upath.relative("/a/b", "/a/b")).toBe("")
            })
        })
    })

    describe("extra functions", () => {
        describe("toUnix", () => {
            it("should convert paths to Unix style", () => {
                expect(upath.toUnix("a\\b\\c")).toBe("a/b/c")
                expect(upath.toUnix("\\a\\b\\c")).toBe("/a/b/c")
                expect(upath.toUnix("a//b//c")).toBe("a/b/c")
                expect(upath.toUnix("//server//share")).toBe("//server/share")
            })
        })

        describe("normalizeSafe", () => {
            it("should normalize paths safely", () => {
                expect(upath.normalizeSafe("./a/b/c")).toBe("./a/b/c")
                expect(upath.normalizeSafe("./a/../b")).toBe("./b")
                expect(upath.normalizeSafe("//server/share")).toBe("//server/share")
                expect(upath.normalizeSafe("//./path")).toBe("//./path")
            })
        })

        describe("normalizeTrim", () => {
            it("should normalize and trim trailing slashes", () => {
                expect(upath.normalizeTrim("a/b/c/")).toBe("a/b/c")
                expect(upath.normalizeTrim("./a/b/c/")).toBe("./a/b/c")
                expect(upath.normalizeTrim("/")).toBe("/")
                expect(upath.normalizeTrim("a/")).toBe("a")
            })
        })

        describe("joinSafe", () => {
            it("should join paths safely", () => {
                expect(upath.joinSafe("./a", "b")).toBe("./a/b")
                expect(upath.joinSafe("//server", "share")).toBe("//server/share")
                expect(upath.joinSafe("a", "../b")).toBe("b")
            })
        })

        describe("addExt", () => {
            it("should add extension correctly", () => {
                expect(upath.addExt("file", "txt")).toBe("file.txt")
                expect(upath.addExt("file", ".txt")).toBe("file.txt")
                expect(upath.addExt("file.txt", "txt")).toBe("file.txt")
                expect(upath.addExt("file.js", "txt")).toBe("file.js.txt")
                expect(upath.addExt("file")).toBe("file")
            })
        })

        describe("trimExt", () => {
            it("should trim valid extensions", () => {
                expect(upath.trimExt("file.txt")).toBe("file")
                expect(upath.trimExt("file.js")).toBe("file")
                expect(upath.trimExt("file.txt", [".txt"])).toBe("file.txt")
                expect(upath.trimExt("file.longext", [], 5)).toBe("file.longext")
                expect(upath.trimExt("file")).toBe("file")
                expect(upath.trimExt(".gitignore")).toBe(".gitignore")
            })
        })

        describe("removeExt", () => {
            it("should remove specific extensions", () => {
                expect(upath.removeExt("file.txt", "txt")).toBe("file")
                expect(upath.removeExt("file.txt", ".txt")).toBe("file")
                expect(upath.removeExt("file.js", "txt")).toBe("file.js")
                expect(upath.removeExt("file.js")).toBe("file.js")
                expect(upath.removeExt("file")).toBe("file")
            })
        })

        describe("changeExt", () => {
            it("should change extensions correctly", () => {
                expect(upath.changeExt("file.txt", "js")).toBe("file.js")
                expect(upath.changeExt("file.txt", ".js")).toBe("file.js")
                expect(upath.changeExt("file", "js")).toBe("file.js")
                expect(upath.changeExt("file.txt")).toBe("file")
                expect(upath.changeExt("file.txt", "js", [".txt"])).toBe("file.txt.js")
            })
        })

        describe("defaultExt", () => {
            it("should add default extension when needed", () => {
                expect(upath.defaultExt("file", "txt")).toBe("file.txt")
                expect(upath.defaultExt("file.js", "txt")).toBe("file.js")
                expect(upath.defaultExt("file", ".txt")).toBe("file.txt")
                expect(upath.defaultExt("file.js", "txt", [".js"])).toBe("file.js.txt")
                expect(upath.defaultExt("file.longext", "txt", [], 5)).toBe("file.longext.txt")
            })
        })
    })
})
