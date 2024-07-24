import { useParams } from 'react-router-dom'
import Form from '~components/PollForm/Form'
import { PollFormProvider } from '~components/PollForm/FormContext'

const AppForm = () => {
  const { id } = useParams<{ id: CommunityID }>()

  return (
    <PollFormProvider>
      <Form communityId={id} />
    </PollFormProvider>
  )
}

export default AppForm
