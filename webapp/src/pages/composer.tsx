import { Composer } from '~components/PollForm/Composer'
import { PollFormProvider } from '~components/PollForm/FormContext'

const ComposerPage = () => (
  <PollFormProvider>
    <Composer />
  </PollFormProvider>
)

export default ComposerPage
