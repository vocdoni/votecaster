import {
  Alert,
  AlertIcon,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  FlexProps,
  Heading,
  VStack,
} from '@chakra-ui/react'
import type { FC } from 'react'
import { ReputationCard } from '~components/Auth/Reputation'
import { SignInButton } from '~components/Auth/SignInButton'
import { useAuth } from '~components/Auth/useAuth'
import CensusTypeSelector from '~components/CensusTypeSelector'
import { useReputation } from '~components/Reputation/useReputation'
import { Choices } from './Choices'
import { Done } from './Done'
import { Duration } from './Duration'
import { Notify } from './Notify'
import { Question } from './Question'
import { usePollForm } from './usePollForm'

type FormProps = FlexProps & {
  communityId?: CommunityID
  composer?: boolean
}

const Form: FC<FormProps> = ({ communityId, composer, ...props }) => {
  const { error, pid, form: methods, onSubmit, loading, status } = usePollForm()
  const { handleSubmit } = methods
  const { reputation } = useReputation()
  const { isAuthenticated, logout } = useAuth()

  return (
    <Flex flexDir='column' alignItems='center' w={{ base: 'full', sm: 450, md: 550 }} {...props}>
      <Card w='100%' borderRadius={composer ? 0 : 6}>
        <CardHeader textAlign='center'>
          <Heading as='h2' size='lg' textAlign='center'>
            Create a framed poll
          </Heading>
        </CardHeader>
        <CardBody>
          <VStack as='form' onSubmit={handleSubmit(onSubmit)} spacing={4} align='left'>
            {pid ? (
              <Done />
            ) : (
              <>
                <Question />
                <Choices />
                <CensusTypeSelector complete isDisabled={loading} composer={composer} communityId={communityId} />
                <Notify />
                <Duration />

                {error && (
                  <Alert status='error'>
                    <AlertIcon />
                    {error}
                  </Alert>
                )}
                {isAuthenticated ? (
                  <>
                    <Button type='submit' isLoading={loading} loadingText={status}>
                      Create
                    </Button>
                    {!composer && (
                      <>
                        <Box fontSize='xs' textAlign='right'>
                          or{' '}
                          <Button variant='text' size='xs' p={0} onClick={logout} height='auto'>
                            logout
                          </Button>
                        </Box>
                        <ReputationCard reputation={reputation!} />
                      </>
                    )}
                  </>
                ) : (
                  <Box display='flex' justifyContent='center' alignItems='center' flexDir='column'>
                    <SignInButton size='lg' />
                    to create a poll
                  </Box>
                )}
              </>
            )}
          </VStack>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default Form
