import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Alert,
  AlertIcon,
  Box,
  Button,
  Text,
  VStack,
} from '@chakra-ui/react'
import React from 'react'
import { TokenSignin } from '~components/Auth/TokenSignIn'
import { useAuth } from '~components/Auth/useAuth'
import CensusTypeSelector from '~components/CensusTypeSelector'
import { Choices } from './Choices'
import { Duration } from './Duration'
import { Notify } from './Notify'
import { Question } from './Question'
import { usePollForm } from './usePollForm'

export const Composer: React.FC = () => {
  const { isAuthenticated } = useAuth()
  const {
    form: { handleSubmit },
    error,
    status,
    onSubmit,
    loading,
  } = usePollForm()

  return (
    <Box as='form' onSubmit={handleSubmit(onSubmit)}>
      <VStack spacing={4} alignItems='start'>
        <Question />
        <Choices />

        {error && (
          <Alert status='error'>
            <AlertIcon />
            {error}
          </Alert>
        )}

        <TokenSignin />
        {isAuthenticated && (
          <Button
            type='submit'
            colorScheme='purple'
            isDisabled={!isAuthenticated}
            isLoading={loading}
            w='full'
            loadingText={status}
          >
            Create poll
          </Button>
        )}
        <Accordion allowToggle>
          <AccordionItem>
            <AccordionButton>
              <AccordionIcon color='purple.500' />
              <Text fontWeight={600}>Advanced settings</Text>
            </AccordionButton>
            <AccordionPanel as={VStack} gap={4}>
              <CensusTypeSelector composer complete showAsSelect isDisabled={loading} />
              <Notify />
              <Duration />
            </AccordionPanel>
          </AccordionItem>
        </Accordion>
      </VStack>
    </Box>
  )
}
