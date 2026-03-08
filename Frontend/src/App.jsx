import React from 'react'
import { Route, Routes } from "react-router"
import toast, { Toaster } from 'react-hot-toast'

const App = () => {
  return (
    <div>
      <Toaster />
      <Routes>
        <Route path="/" element={<div>Home</div>} />
      </Routes>
    </div>
  )
}

export default App