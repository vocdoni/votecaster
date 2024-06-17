import { Avatar, Box, Button, Card, CardBody, Flex, Heading, Icon, Link } from '@chakra-ui/react'
import { useFormContext } from 'react-hook-form'
import { FaExternalLinkAlt } from 'react-icons/fa'
import { MdHowToVote } from 'react-icons/md'
import { Link as RouterLink } from 'react-router-dom'
import { useAccount } from 'wagmi'
import { chainExplorer } from '~util/chain'
import { CommunityMetaFormValues } from './Meta'

type DoneProps = {
  tx: string
  id: string | null
}

const CommunityDone = ({ tx, id }: DoneProps) => {
  const { watch } = useFormContext<CommunityMetaFormValues>()
  const { chain } = useAccount()
  const src = watch('src')

  return (
    <Flex flexDir='column' alignItems='center' w={{ base: 'full', sm: 450, md: 500 }}>
      <Card w='100%'>
        <CardBody my={10}>
          <Flex direction={'column'} justifyItems={'center'} textAlign={'center'} gap={6}>
            {src && (
              <Box>
                <Avatar src={src} size={'xl'} />
              </Box>
            )}
            <Heading mb={10} size='lg'>
              Your community is now live on
              <Link href={`${chainExplorer(chain)}/tx/${tx}`} isExternal>
                {' '}
                {chain?.name}
                <Icon as={FaExternalLinkAlt} w={4} />
              </Link>
            </Heading>
            <Heading size='md'>
              Get started by creating polls
              <br />
              to engage with your members!
            </Heading>
            <RouterLink to={`/form/${id}`}>
              <Button leftIcon={<MdHowToVote />}>Create your first vote</Button>
            </RouterLink>
          </Flex>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default CommunityDone
