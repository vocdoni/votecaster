import { useParams } from 'react-router-dom'
import Form from '~components/Form'

const AppForm = () => {
  const { id } = useParams<{ id: CommunityID }>()
  return <Form communityId={id} />
}

export default AppForm
