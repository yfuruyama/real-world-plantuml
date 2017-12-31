package com.yfuruyama.syntaxchecker

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.SerializationFeature
import com.fasterxml.jackson.module.kotlin.KotlinModule
import net.sourceforge.plantuml.syntax.SyntaxChecker
import org.glassfish.jersey.jackson.JacksonFeature
import org.glassfish.jersey.jetty.JettyHttpContainerFactory
import org.glassfish.jersey.server.ResourceConfig
import java.util.logging.Logger
import javax.ws.rs.*
import javax.ws.rs.core.MediaType
import javax.ws.rs.core.UriBuilder
import javax.ws.rs.ext.ContextResolver
import javax.ws.rs.ext.Provider

fun main(args: Array<String>) {
    val port = 8087
    println("Starts server: port=%d".format(port))

    val baseUri = UriBuilder.fromUri("http://localhost/").port(port).build()
    val config = ResourceConfig()
            .register(JacksonFeature::class.java)
            .register(ObjectMapperProvider::class.java)
            .register(CheckSyntaxResource())

    val server = JettyHttpContainerFactory.createServer(baseUri, config)
    try {
        server.join()
    } finally {
        server.destroy()
    }
}

@Provider
class ObjectMapperProvider : ContextResolver<ObjectMapper> {
    val objectMapper = ObjectMapper()
        .enable(SerializationFeature.INDENT_OUTPUT)
        .registerModule(KotlinModule())

    override fun getContext(type: Class<*>?): ObjectMapper? = objectMapper
}

@JsonIgnoreProperties(ignoreUnknown = true)
data class CheckSyntaxRequest(val source: String?)

data class CheckSyntaxResponse(val valid: Boolean, val diagramType: String, val description: String)

@Path("check_syntax")
class CheckSyntaxResource {
    var logger = Logger.getLogger(CheckSyntaxResource::class.java.name)

    @POST
    @Produces(MediaType.APPLICATION_JSON)
    @Consumes(MediaType.APPLICATION_JSON)
    fun checkSyntax(req: CheckSyntaxRequest): CheckSyntaxResponse {
        if (req.source == null) {
            throw BadRequestException("`source` not specified")
        }

        logger.info("Get source %s".format(req.source))
        val result = SyntaxChecker.checkSyntax(req.source)

        if (result.isError || result.umlDiagramType == null) {
            logger.info("Invalid syntax: errors=%s".format(result.errors.joinToString(",")))
            return CheckSyntaxResponse(false, "", "")
        } else {
            logger.info("Valid syntax: diagramType=%s, description=%s".format(result.umlDiagramType, result.description))
            return CheckSyntaxResponse(true, result.umlDiagramType.name, result.description)
        }
    }
}
