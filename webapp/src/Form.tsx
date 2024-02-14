import {
  Alert,
  AlertIcon,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  FlexProps,
  FormControl,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  Heading,
  Image,
  Input,
  Link,
  Text,
  VStack,
} from '@chakra-ui/react'
import { useProfile } from '@farcaster/auth-kit'
import axios from 'axios'
import React, { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Done } from './Done'
import logo from '/logo.svg'

interface FormValues {
  question: string
  choice1: string
  choice2: string
  choice3?: string
  choice4?: string
  duration?: number
}

const appUrl = import.meta.env.APP_URL

const Form: React.FC = (props: FlexProps) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>()
  const { profile } = useProfile()
  const [loading, setLoading] = useState<boolean>(false)
  const [pid, setPid] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const onSubmit = async (data: FormValues) => {
    setError(null)
    try {
      setLoading(true)
      const election = {
        question: data.question,
        duration: Number(data.duration),
        options: [],
        profile,
      }

      for (let i = 1; i < 5; i++) {
        if (data[`choice${i}`]) {
          election.options.push(data[`choice${i}`])
        }
      }

      const res = await axios.post(`${appUrl}/create`, election)
      setPid(res.data.replace('\n', ''))
    } catch (e) {
      console.error('there was an error creating the election:', e)
      if ('message' in e) {
        setError(e.message)
      }
    } finally {
      setLoading(false)
    }
  }

  const required = {
    value: true,
    message: 'This field is required',
  }
  const maxLength = {
    value: 30,
    message: 'Max length is 30 characters',
  }

  return (
    <Flex flexDir='column' alignItems='center' {...props}>
      <Card maxW={500}>
        <CardHeader align='center'>
          <Image src={logo} alt='farcaster.vote logo' mb={4} />
          <Heading as='h1' size='lg' textAlign='center'>
            Create a new farcaster voting
          </Heading>
        </CardHeader>
        <CardBody>
          <VStack as='form' onSubmit={handleSubmit(onSubmit)} spacing={4} align='left'>
            {pid ? (
              <Done pid={pid} />
            ) : (
              <>
                <FormControl isRequired isDisabled={loading} isInvalid={!!errors.question}>
                  <FormLabel htmlFor='question'>Question</FormLabel>
                  <Input
                    id='question'
                    placeholder='Enter your question'
                    {...register('question', {
                      required,
                      maxLength: { value: 250, message: 'Max length is 250 characters' },
                    })}
                  />
                  <FormErrorMessage>{errors.question?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <FormControl isRequired isDisabled={loading} isInvalid={!!errors.choice1}>
                  <FormLabel htmlFor='choice1'>Choice 1</FormLabel>
                  <Input id='choice1' placeholder='Enter choice 1' {...register('choice1', { required, maxLength })} />
                  <FormErrorMessage>{errors.choice1?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <FormControl isRequired isDisabled={loading} isInvalid={!!errors.choice2}>
                  <FormLabel htmlFor='choice2'>Choice 2</FormLabel>
                  <Input id='choice2' placeholder='Enter choice 2' {...register('choice2', { required, maxLength })} />
                  <FormErrorMessage>{errors.choice2?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <FormControl isDisabled={loading} isInvalid={!!errors.choice3}>
                  <FormLabel htmlFor='choice3'>Choice 3 (Optional)</FormLabel>
                  <Input id='choice3' placeholder='Enter choice 3 (Optional)' {...register('choice3', { maxLength })} />
                  <FormErrorMessage>{errors.choice3?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <FormControl isDisabled={loading} isInvalid={!!errors.choice4}>
                  <FormLabel htmlFor='choice4'>Choice 4 (Optional)</FormLabel>
                  <Input id='choice4' placeholder='Enter choice 4 (Optional)' {...register('choice4', { maxLength })} />
                  <FormErrorMessage>{errors.choice4?.message?.toString()}</FormErrorMessage>
                </FormControl>
                <FormControl isDisabled={loading} isInvalid={!!errors.duration}>
                  <FormLabel htmlFor='duration'>Duration (Optional)</FormLabel>
                  <Input
                    id='duration'
                    placeholder='Enter duration (in hours)'
                    {...register('duration')}
                    type='number'
                    min={1}
                  />
                  <FormErrorMessage>{errors.duration?.message?.toString()}</FormErrorMessage>
                  <FormHelperText>24h by default</FormHelperText>
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
      <Text mt={3} fontSize='.8em' textAlign='center'>
        <Link href='https://warpcast.com/vocdoni' target='_blank'>
          By @vocdoni
        </Link>
      </Text>
    </Flex>
  )
}

export default Form
