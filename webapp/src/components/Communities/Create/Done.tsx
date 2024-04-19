type DoneProps = {
  tx: string
  communityId?: string | null
}

const CommunityDone = ({tx, communityId}: DoneProps) => {
  return (
    <div>
      <h1>Created Succesfully</h1>
      <p>Community ID: {communityId}</p>
      <p>Transaction ID: {tx}</p>
    </div>
  )
}

export default CommunityDone