import { useContext } from 'react'
import { PollFormContext } from './FormContext'

export const usePollForm = () => {
  const ctxt = useContext(PollFormContext)
  if (!ctxt) {
    throw new Error(
      'usePollForm returned `undefined`, maybe you forgot to wrap the component within <PollFormProvider />?'
    )
  }

  return ctxt
}
