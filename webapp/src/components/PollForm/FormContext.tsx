import { createContext, PropsWithChildren } from 'react'
import { FormProvider } from 'react-hook-form'
import { PollFormProviderProps, usePollFormProvider } from './usePollFormProvider'

export type PollFormState = ReturnType<typeof usePollFormProvider>

export const PollFormContext = createContext<PollFormState | undefined>(undefined)

export type PollFormProviderComponentProps = PollFormProviderProps

export const PollFormProvider = ({ children }: PropsWithChildren<PollFormProviderComponentProps>) => {
  const value = usePollFormProvider()

  return (
    <PollFormContext.Provider value={value}>
      <FormProvider {...value.form}>{children}</FormProvider>
    </PollFormContext.Provider>
  )
}
PollFormProvider.displayName = 'PollFormProvider'
