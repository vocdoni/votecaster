import {
  Alert,
  AlertDescription,
  Button,
  Flex,
  FormControl,
  FormControlProps,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputRightElement,
  Link,
  ListItem,
  Radio,
  RadioGroup,
  Select,
  Stack,
  UnorderedList,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useFieldArray, useFormContext } from 'react-hook-form'
import { BiTrash } from 'react-icons/bi'
import Airstack from '../assets/airstack.svg?react'
import { fetchAirstackBlockchains } from '../queries/census'
import { appUrl } from '../util/constants'
import { cleanChannel, ucfirst } from '../util/strings'
import { useAuth } from './Auth/useAuth'

export type CensusType = 'farcaster' | 'channel' | 'followers' | 'custom' | 'erc20' | 'nft'

export type Address = {
  address: string
  blockchain: string
}

export type CensusFormValues = {
  censusType: CensusType
  addresses?: Address[]
  channel?: string
  csv?: File | undefined
}

const CensusTypeSelector = ({ withCSV, ...props }: FormControlProps & { withCSV?: boolean }) => {
  const { bfetch } = useAuth()
  const {
    watch,
    formState: { errors },
    setValue,
    control,
    register,
  } = useFormContext<CensusFormValues>()
  const {
    fields: addressFields,
    append: appendAddress,
    remove: removeAddress,
  } = useFieldArray({
    control,
    name: 'addresses',
  })
  const { data: blockchains, isLoading } = useQuery({
    queryKey: ['blockchains'],
    queryFn: fetchAirstackBlockchains(bfetch),
  })

  const censusType = watch('censusType')

  // reset address fields when censusType changes
  useEffect(() => {
    if (censusType === 'erc20' || censusType === 'nft') {
      // Remove all fields initially
      setValue('addresses', [])
      // Add one field by default
      for (let i = 0; i < 1; i++) {
        appendAddress({ address: '', blockchain: 'base' })
      }
    }
  }, [censusType, appendAddress, removeAddress])

  const required = {
    value: true,
    message: 'This field is required',
  }

  return (
    <>
      <FormControl {...props}>
        <FormLabel>Census/voters</FormLabel>
        <RadioGroup onChange={(val: string) => setValue('censusType', val)} value={censusType} id='census-type'>
          <Stack direction='column' flexWrap='wrap'>
            <Radio value='farcaster'>🌐 All farcaster users</Radio>
            <Radio value='channel'>⛩ Channel gated</Radio>
            <Radio value='followers'>❤️ My followers and me</Radio>
            {withCSV && <Radio value='custom'>🦄 Token based via CSV</Radio>}
            <Radio value='nft'>
              <Icon as={Airstack} /> NFT based via airstack
            </Radio>
            <Radio value='erc20'>
              <Icon as={Airstack} /> ERC20 based via airstack
            </Radio>
          </Stack>
        </RadioGroup>
      </FormControl>
      {['erc20', 'nft'].includes(censusType) &&
        addressFields.map((field, index) => (
          <FormControl key={field.id} {...props}>
            <FormLabel>
              {censusType.toUpperCase()} address {index + 1}
            </FormLabel>
            <Flex>
              <Select
                {...register(`addresses.${index}.blockchain`, { required })}
                defaultValue='ethereum'
                w='auto'
                _loading={isLoading}
              >
                {blockchains.map((blockchain, key) => (
                  <option value={blockchain} key={key}>
                    {ucfirst(blockchain)}
                  </option>
                ))}
              </Select>
              <InputGroup>
                <Input placeholder='Smart contract address' {...register(`addresses.${index}.address`, { required })} />
                {addressFields.length > 1 && (
                  <InputRightElement>
                    <IconButton
                      aria-label='Remove address'
                      icon={<BiTrash />}
                      onClick={() => removeAddress(index)}
                      size='sm'
                    />
                  </InputRightElement>
                )}
              </InputGroup>
            </Flex>
          </FormControl>
        ))}
      {censusType === 'nft' && addressFields.length < 3 && (
        <Button variant='ghost' onClick={() => appendAddress({ address: '' })}>
          Add address
        </Button>
      )}
      {censusType === 'channel' && (
        <FormControl isRequired isInvalid={!!errors.channel} {...props}>
          <FormLabel htmlFor='channel'>Channel slug (URL identifier)</FormLabel>
          <Input
            id='channel'
            placeholder='Enter channel i.e. degen'
            {...register('channel', {
              required,
              validate: async (val) => {
                val = cleanChannel(val)
                try {
                  const res = await bfetch(`${appUrl}/census/channel-gated/${val}/exists`)
                  if (res.status === 200) {
                    return true
                  }
                } catch (e) {
                  return 'Invalid channel specified'
                }
                return 'Invalid channel specified'
              },
            })}
          />
          <FormErrorMessage>{errors.channel?.message?.toString()}</FormErrorMessage>
        </FormControl>
      )}
      {censusType === 'custom' && (
        <FormControl isRequired {...props}>
          <FormLabel htmlFor='csv'>CSV files</FormLabel>
          <Input
            id='csv'
            placeholder='Upload CSV'
            type='file'
            multiple
            accept='text/csv,application/csv,.csv'
            {...register('csv', {
              required: {
                value: true,
                message: 'This field is required',
              },
            })}
          />
          {errors.csv ? (
            <FormErrorMessage>{errors.csv?.message?.toString()}</FormErrorMessage>
          ) : (
            <FormHelperText>
              <Alert status='info'>
                <AlertDescription>
                  The CSV files <strong>must include Ethereum addresses and their balances</strong> from any network.
                  You can build your own at:
                  <UnorderedList>
                    <ListItem>
                      <Link isExternal href='https://holders.at' variant='primary'>
                        holders.at
                      </Link>{' '}
                      for NFTs
                    </ListItem>
                    <ListItem>
                      <Link isExternal href='https://collectors.poap.xyz' variant='primary'>
                        collectors.poap.xyz
                      </Link>{' '}
                      for POAPs
                    </ListItem>
                  </UnorderedList>
                  <strong>If an address appears multiple times, its balances will be aggregated.</strong>
                </AlertDescription>
              </Alert>
            </FormHelperText>
          )}
        </FormControl>
      )}
    </>
  )
}

export default CensusTypeSelector
