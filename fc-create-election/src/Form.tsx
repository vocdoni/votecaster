import {
  Alert,
  AlertIcon,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  FormControl,
  FormLabel,
  Heading,
  Image,
  Input,
  Link,
  Text,
  VStack,
} from '@chakra-ui/react'
import axios from 'axios'
import React, { useState } from 'react'
import { useForm } from 'react-hook-form'

interface FormValues {
  question: string
  choice1: string
  choice2: string
  choice3?: string
  choice4?: string
}

const Form: React.FC = () => {
  const {
    register,
    handleSubmit,
    // formState: { errors },
  } = useForm<FormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const backendUrl = import.meta.env.BACKEND_URL

  const onSubmit = async (data: FormValues) => {
    setError(null)
    try {
      setLoading(true)
      const election = {
        question: data.question,
        options: [],
      }
      for (let i = 1; i < 5; i++) {
        if (data[`choice${i}`]) {
          election.options.push(data[`choice${i}`])
        }
      }

      const res = await axios.post(`${backendUrl}/create`, election)
      setPid(res.data.replace('\n', ''))
    } catch (e) {
      if ('message' in e) {
        setError(e.message)
      } else {
        console.error('there was an error creating the election:', e)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <Flex minH='100vh' justifyContent='center' alignItems='center'>
      <Card>
        <CardHeader align='center'>
          <Image
            src='https://assets-global.website-files.com/6398d7c1bcc2b775ebaa4f2f/6398d7c1bcc2b75440aa4f50_vocdoni-imagotype.svg'
            alt='Logo'
            mb={4}
          />
          <Heading as='h1' size='lg' textAlign='center'>
            Create a new farcaster voting
          </Heading>
        </CardHeader>
        <CardBody>
          <VStack as='form' onSubmit={handleSubmit(onSubmit)} spacing={4} align='stretch'>
            {pid ? (
              <Text display='inline'>
                Done! You can now share it using this link:
                <br />
                <Link href={`https://farcaster.vote/${pid}`}>https://farcaster.vote/{pid}</Link>
              </Text>
            ) : (
              <>
                <FormControl isRequired isDisabled={loading}>
                  <FormLabel htmlFor='question'>Question</FormLabel>
                  <Input
                    id='question'
                    placeholder='Enter your question'
                    {...register('question', { required: true })}
                  />
                </FormControl>
                <FormControl isRequired isDisabled={loading}>
                  <FormLabel htmlFor='choice1'>Choice 1</FormLabel>
                  <Input id='choice1' placeholder='Enter choice 1' {...register('choice1', { required: true })} />
                </FormControl>
                <FormControl isRequired isDisabled={loading}>
                  <FormLabel htmlFor='choice2'>Choice 2</FormLabel>
                  <Input id='choice2' placeholder='Enter choice 2' {...register('choice2', { required: true })} />
                </FormControl>
                <FormControl isDisabled={loading}>
                  <FormLabel htmlFor='choice3'>Choice 3 (Optional)</FormLabel>
                  <Input id='choice3' placeholder='Enter choice 3 (Optional)' {...register('choice3')} />
                </FormControl>
                <FormControl isDisabled={loading}>
                  <FormLabel htmlFor='choice4'>Choice 4 (Optional)</FormLabel>
                  <Input id='choice4' placeholder='Enter choice 4 (Optional)' {...register('choice4')} />
                </FormControl>
                {error && (
                  <Alert status='error'>
                    <AlertIcon />
                    {error}
                  </Alert>
                )}
                <Button type='submit' colorScheme='purple' isLoading={loading}>
                  Create
                </Button>
              </>
            )}
          </VStack>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default Form
