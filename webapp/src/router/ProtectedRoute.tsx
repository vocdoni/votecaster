import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../components/Auth/useAuth'

const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth()

  return isAuthenticated ? <Outlet /> : <Navigate to='/' replace={true} />
}

export default ProtectedRoute
