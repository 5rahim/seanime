"use client" //  It marks this as a browser-only component.

import React, { useEffect } from 'react'

export function KeyboardShortcuts() {
  useEffect(() => {
    // This function will be called whenever a key is pressed
    const handleKeyDown = (event: KeyboardEvent) => {
      // Check for Alt + Left Arrow for "Back"
      if (event.altKey && event.key === 'ArrowLeft') {
        window.history.back()
      }
      // Check for Alt + Right Arrow for "Forward"
      if (event.altKey && event.key === 'ArrowRight') {
        window.history.forward()
      }
    }

    // Attach our function to the window's keydown event
    window.addEventListener('keydown', handleKeyDown)

    // This is a cleanup function. It removes the event listener
    // when the component is no longer needed, preventing bugs.
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, []) // The empty array [] ensures this setup code only runs once.

  // This component does not render any visible HTML, so it returns null.
  return null
}