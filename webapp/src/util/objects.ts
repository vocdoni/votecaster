const findDelegationPath = (delegations: Delegation[], start: number): number[] => {
  const path: number[] = [start]

  const nextDelegation = delegations.find((d) => d.from === start)

  if (nextDelegation) {
    path.push(...findDelegationPath(delegations, nextDelegation.to))
  }

  return path
}

/**
 * Finds all possible delegation paths in a given array of delegations.
 * @param delegations - An array of delegation objects.
 * @returns An array of arrays, each representing a delegation path.
 */
export const getDelegationsPath = (delegations: Delegation[]): number[][] => {
  const from = new Set(delegations.map((d) => d.from))
  const to = new Set(delegations.map((d) => d.to))

  // Starting points are those 'from' which are not present in 'to' (no one delegates to them)
  const startPoints = [...from].filter((from) => !to.has(from))

  // Find paths for each starting point
  const paths = startPoints.map((start) => findDelegationPath(delegations, start))

  return paths
}
