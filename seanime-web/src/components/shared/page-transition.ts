export const PAGE_TRANSITION = {
    initial: { opacity: 0, y: 60 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: 60 },
    transition: {
        type: "spring",
        damping: 20,
        stiffness: 100,
    },
}
