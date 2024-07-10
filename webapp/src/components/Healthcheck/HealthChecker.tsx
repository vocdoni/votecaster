import React, { useEffect } from 'react'
import useChainHealth from './use-healthcheck'

const HealthChecker: React.FC<{ chainKey: ChainKey; updateStatus: (key: ChainKey, isConnected: boolean) => void }> = ({
  chainKey,
  updateStatus,
}) => {
  const { isConnected } = useChainHealth(chainKey)

  useEffect(() => {
    updateStatus(chainKey, isConnected)
  }, [isConnected, chainKey, updateStatus])

  return null
}

export default HealthChecker
