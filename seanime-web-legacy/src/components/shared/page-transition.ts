export const PAGE_TRANSITION = {
    initial: { opacity: 0, y: 0 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: 0 },
    transition: {
        // duration: 0.3,
        // delay: 0.1,
        type: "spring",
        damping: 20,
        stiffness: 100,
    },
}
