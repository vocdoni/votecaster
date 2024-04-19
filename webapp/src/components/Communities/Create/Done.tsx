import {Box, Flex, Heading, Card, CardBody, Button, Avatar, Link, Icon} from "@chakra-ui/react";
import {MdHowToVote} from "react-icons/md";
import {useFormContext} from "react-hook-form";
import {CommunityMetaFormValues} from "./Meta.tsx";
import {FaExternalLinkAlt} from "react-icons/fa";
import {useNavigate} from 'react-router-dom';

type DoneProps = {
  tx: string
}

const CommunityDone = ({tx}: DoneProps) => {
  const {
    watch,
  } = useFormContext<CommunityMetaFormValues>()
  const logo = watch('logo')
  const navigate = useNavigate(); // Hook to control navigation

  return (
    <Flex flexDir='column' alignItems='center' w={{base: 'full', sm: 450, md: 500}}>
      <Card w='100%'>
        <CardBody my={10}>
          <Flex direction={'column'} justifyItems={'center'} textAlign={'center'} gap={6}>
            {logo && <Box>
              <Avatar src={logo} size={'xl'}/>
            </Box>}
            <Heading mb={10} size='lg'>Your community is now live on
              <Link href={`https://explorer.degen.tips/tx/${tx}`} isExternal> 🎩 Degenchain!
                <Icon as={FaExternalLinkAlt} w={4}/>
              </Link>
            </Heading>
            <Heading size='md'>Get started by creating polls<br/>
              to engage with your members!</Heading>
            <Button onClick={() => navigate('/')} leftIcon={<MdHowToVote/>}>Create your first vote</Button>
          </Flex>
        </CardBody>
      </Card>
    </Flex>
  )
}

export default CommunityDone