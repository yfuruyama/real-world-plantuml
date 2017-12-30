package com.yfuruyama.syntaxchecker

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.SerializationFeature
import com.fasterxml.jackson.module.kotlin.KotlinModule
import net.sourceforge.plantuml.syntax.SyntaxChecker
import org.glassfish.jersey.jackson.JacksonFeature
import org.glassfish.jersey.jetty.JettyHttpContainerFactory
import org.glassfish.jersey.server.ResourceConfig
import javax.ws.rs.*
import javax.ws.rs.core.MediaType
import javax.ws.rs.core.UriBuilder
import javax.ws.rs.ext.ContextResolver
import javax.ws.rs.ext.Provider

fun main(args: Array<String>) {
    val port = 8080
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

data class CheckSyntaxRequest(val source: String)
data class CheckSyntaxResponse(val valid: Boolean, val diagramType: String)

@Path("check_syntax")
class CheckSyntaxResource {
    @POST
    @Produces(MediaType.APPLICATION_JSON)
    @Consumes(MediaType.APPLICATION_JSON)
    fun checkSyntax(req: CheckSyntaxRequest): CheckSyntaxResponse {
        // TODO: body check
        println("Get source %s".format(req.source))
        val result = SyntaxChecker.checkSyntax(req.source)
        // TODO: log
        println("Syntax check result: %s".format(result))

        // TODO: UNKNOWN diagram type
        return CheckSyntaxResponse(!result.isError, result.umlDiagramType.name)
    }
}
